package tests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"usuario-service/internal/domain/entity"
	domainerrors "usuario-service/internal/domain/errors"
	"usuario-service/internal/domain/valueobject"
	"usuario-service/internal/infrastructure/database"
	"usuario-service/internal/infrastructure/persistence"
)

func setupTestDB(t *testing.T) (*gorm.DB, *persistence.GormUsuarioRepository) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "korp"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "korp123"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "usuario_test_db"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := database.NewPostgresConnection(dsn)
	if err != nil {
		t.Skipf("Requires PostgreSQL, skipping test: %v", err)
		return nil, nil
	}

	err = db.AutoMigrate(&persistence.UsuarioModel{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db.Exec("DELETE FROM usuarios")

	repo := persistence.NewGormUsuarioRepository(db)
	return db, repo
}

func TestConcurrentUpdateSameUsuario(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	t.Cleanup(func() { db.Exec("DELETE FROM usuarios") })

	ctx := context.Background()
	id := uuid.New().String()
	u, err := entity.NewUsuario(id, "Test User", fmt.Sprintf("test-%s@test.com", id), "52998224725")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	var wg sync.WaitGroup
	var successCount int32
	var failCount int32

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			userToUpdate, err := repo.FindByID(ctx, id)
			if err != nil {
				return
			}

			err = userToUpdate.AtualizarNome(fmt.Sprintf("Updated Name %d", idx))
			if err != nil {
				return
			}

			userToUpdate.IncrementVersion()

			err = repo.Update(ctx, userToUpdate)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else if errors.Is(err, domainerrors.ErrConcurrencyConflict) {
				atomic.AddInt32(&failCount, 1)
			}
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful update, got %d", successCount)
	}
	if failCount != 9 {
		t.Errorf("Expected exactly 9 failed updates due to concurrency, got %d", failCount)
	}

	finalUser, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}

	if finalUser.Version() != 2 {
		t.Errorf("Expected version to be 2, got %d", finalUser.Version())
	}
}

func TestConcurrentCreateSameEmail(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	t.Cleanup(func() { db.Exec("DELETE FROM usuarios") })

	ctx := context.Background()
	var wg sync.WaitGroup
	var successCount int32
	var failCount int32

	email := fmt.Sprintf("sameemail-%s@test.com", uuid.New().String())
	cpfs := []string{"52998224725", "11144477735", "36243289041", "77736490070", "85596975006"}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			id := uuid.New().String()
			u, err := entity.NewUsuario(id, fmt.Sprintf("User %d", idx), email, cpfs[idx])
			if err != nil {
				return
			}

			err = repo.Save(ctx, u)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failCount, 1)
			}
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful create, got %d", successCount)
	}
	if failCount != 4 {
		t.Errorf("Expected exactly 4 failed creates, got %d", failCount)
	}

	emailVO, _ := valueobject.NewEmail(email)
	foundUser, err := repo.FindByEmail(ctx, emailVO)
	if err != nil {
		t.Fatalf("Failed to find user by email: %v", err)
	}
	if foundUser == nil {
		t.Errorf("Expected exactly 1 user in DB")
	}
}

func TestConcurrentCreateSameCPF(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	t.Cleanup(func() { db.Exec("DELETE FROM usuarios") })

	ctx := context.Background()
	var wg sync.WaitGroup
	var successCount int32
	var failCount int32

	cpf := "52998224725"

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			id := uuid.New().String()
			email := fmt.Sprintf("diffemail-%d-%s@test.com", idx, uuid.New().String())
			u, err := entity.NewUsuario(id, fmt.Sprintf("User %d", idx), email, cpf)
			if err != nil {
				return
			}

			err = repo.Save(ctx, u)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failCount, 1)
			}
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful create, got %d", successCount)
	}
	if failCount != 4 {
		t.Errorf("Expected exactly 4 failed creates, got %d", failCount)
	}
}

func TestConcurrentDeleteAndUpdate(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	t.Cleanup(func() { db.Exec("DELETE FROM usuarios") })

	ctx := context.Background()
	id := uuid.New().String()
	u, err := entity.NewUsuario(id, "Test User", fmt.Sprintf("test-%s@test.com", id), "52998224725")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	var wg sync.WaitGroup
	var successCount int32
	var failCount int32

	wg.Add(2)

	go func() {
		defer wg.Done()
		user, err := repo.FindByID(ctx, id)
		if err != nil {
			return
		}

		err = user.Desativar()
		if err != nil {
			return
		}
		user.IncrementVersion()

		err = repo.Update(ctx, user)
		if err == nil {
			atomic.AddInt32(&successCount, 1)
		} else if errors.Is(err, domainerrors.ErrConcurrencyConflict) {
			atomic.AddInt32(&failCount, 1)
		}
	}()

	go func() {
		defer wg.Done()
		user, err := repo.FindByID(ctx, id)
		if err != nil {
			return
		}

		err = user.AtualizarNome("Updated")
		if err != nil {
			return
		}
		user.IncrementVersion()

		err = repo.Update(ctx, user)
		if err == nil {
			atomic.AddInt32(&successCount, 1)
		} else if errors.Is(err, domainerrors.ErrConcurrencyConflict) {
			atomic.AddInt32(&failCount, 1)
		}
	}()

	wg.Wait()

	if successCount != 1 {
		t.Errorf("Expected exactly 1 success, got %d", successCount)
	}
	if failCount != 1 {
		t.Errorf("Expected exactly 1 failure, got %d", failCount)
	}
}

func TestOptimisticLockVersionIncrement(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	t.Cleanup(func() { db.Exec("DELETE FROM usuarios") })

	ctx := context.Background()
	id := uuid.New().String()
	u, err := entity.NewUsuario(id, "Initial Name", fmt.Sprintf("test-%s@test.com", id), "52998224725")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	userStale, _ := repo.FindByID(ctx, id)

	user1, _ := repo.FindByID(ctx, id)
	err = user1.AtualizarNome("New Name")
	if err != nil {
		t.Fatalf("Failed to update name: %v", err)
	}
	user1.IncrementVersion()
	err = repo.Update(ctx, user1)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	err = userStale.AtualizarNome("Another Name")
	if err != nil {
		t.Fatalf("Failed to update name: %v", err)
	}
	userStale.IncrementVersion()
	err = repo.Update(ctx, userStale)
	if !errors.Is(err, domainerrors.ErrConcurrencyConflict) {
		t.Errorf("Expected ErrConcurrencyConflict, got %v", err)
	}

	user2, _ := repo.FindByID(ctx, id)
	err = user2.AtualizarNome("Correct Name")
	if err != nil {
		t.Fatalf("Failed to update name: %v", err)
	}
	user2.IncrementVersion()
	err = repo.Update(ctx, user2)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	finalUser, _ := repo.FindByID(ctx, id)
	if finalUser.Version() != 3 {
		t.Errorf("Expected version 3, got %d", finalUser.Version())
	}
}
