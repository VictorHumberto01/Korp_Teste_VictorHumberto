package tests

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"usuario-service/internal/domain/entity"
	domainerrors "usuario-service/internal/domain/errors"
	"usuario-service/internal/domain/valueobject"
)

func generateValidCPF(base int) string {
	cpf := make([]int, 11)
	for i := 0; i < 9; i++ {
		cpf[8-i] = base % 10
		base /= 10
	}

	sum1 := 0
	for i := 0; i < 9; i++ {
		sum1 += cpf[i] * (10 - i)
	}
	rem1 := sum1 % 11
	if rem1 < 2 {
		cpf[9] = 0
	} else {
		cpf[9] = 11 - rem1
	}

	sum2 := 0
	for i := 0; i < 10; i++ {
		sum2 += cpf[i] * (11 - i)
	}
	rem2 := sum2 % 11
	if rem2 < 2 {
		cpf[10] = 0
	} else {
		cpf[10] = 11 - rem2
	}

	res := ""
	for _, d := range cpf {
		res += fmt.Sprintf("%d", d)
	}
	return res
}

func TestCreateAndFindUsuario(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	t.Cleanup(func() { db.Exec("DELETE FROM usuarios") })

	ctx := context.Background()
	id := uuid.New().String()
	email := fmt.Sprintf("test-%s@test.com", id)
	cpf := "52998224725"

	u, err := entity.NewUsuario(id, "Integration Test User", email, cpf)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	foundID, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Errorf("FindByID failed: %v", err)
	}
	if foundID.ID() != id {
		t.Errorf("Expected ID %s, got %s", id, foundID.ID())
	}

	emailVO, _ := valueobject.NewEmail(email)
	foundEmail, err := repo.FindByEmail(ctx, emailVO)
	if err != nil {
		t.Errorf("FindByEmail failed: %v", err)
	}
	if foundEmail.Email().Value() != email {
		t.Errorf("Expected Email %s, got %s", email, foundEmail.Email().Value())
	}

	cpfVO, _ := valueobject.NewCPF(cpf)
	foundCPF, err := repo.FindByCPF(ctx, cpfVO)
	if err != nil {
		t.Errorf("FindByCPF failed: %v", err)
	}
	if foundCPF.CPF().Value() != cpfVO.Value() {
		t.Errorf("Expected CPF %s, got %s", cpfVO.Value(), foundCPF.CPF().Value())
	}
}

func TestFindAllWithPagination(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	t.Cleanup(func() { db.Exec("DELETE FROM usuarios") })

	ctx := context.Background()

	for i := 0; i < 15; i++ {
		id := uuid.New().String()
		email := fmt.Sprintf("test-pagination-%d-%s@test.com", i, id)
		cpf := generateValidCPF(123456789 + i)
		u, _ := entity.NewUsuario(id, fmt.Sprintf("User %d", i), email, cpf)
		repo.Save(ctx, u)
	}

	usersPage1, total1, err := repo.FindAll(ctx, 1, 10)
	if err != nil {
		t.Fatalf("FindAll page 1 failed: %v", err)
	}
	if len(usersPage1) != 10 {
		t.Errorf("Expected 10 users on page 1, got %d", len(usersPage1))
	}
	if total1 != 15 {
		t.Errorf("Expected total 15, got %d", total1)
	}

	usersPage2, total2, err := repo.FindAll(ctx, 2, 10)
	if err != nil {
		t.Fatalf("FindAll page 2 failed: %v", err)
	}
	if len(usersPage2) != 5 {
		t.Errorf("Expected 5 users on page 2, got %d", len(usersPage2))
	}
	if total2 != 15 {
		t.Errorf("Expected total 15, got %d", total2)
	}
}

func TestUpdateUsuario(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	t.Cleanup(func() { db.Exec("DELETE FROM usuarios") })

	ctx := context.Background()
	id := uuid.New().String()
	u, _ := entity.NewUsuario(id, "Old Name", fmt.Sprintf("test-%s@test.com", id), "52998224725")
	repo.Save(ctx, u)

	userToUpdate, _ := repo.FindByID(ctx, id)
	userToUpdate.AtualizarNome("New Name")
	userToUpdate.IncrementVersion()
	err := repo.Update(ctx, userToUpdate)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(ctx, id)
	if found.Nome().Value() != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", found.Nome().Value())
	}
}

func TestDeleteUsuario(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	t.Cleanup(func() { db.Exec("DELETE FROM usuarios") })

	ctx := context.Background()
	id := uuid.New().String()
	u, _ := entity.NewUsuario(id, "To Be Deleted", fmt.Sprintf("test-%s@test.com", id), "52998224725")
	repo.Save(ctx, u)

	err := repo.Delete(ctx, id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.FindByID(ctx, id)
	if !errors.Is(err, domainerrors.ErrUsuarioNotFound) {
		t.Errorf("Expected ErrUsuarioNotFound, got %v", err)
	}

	users, _, _ := repo.FindAll(ctx, 1, 10)
	for _, user := range users {
		if user.ID() == id {
			t.Errorf("Deleted user should not appear in FindAll")
		}
	}
}

func TestEmailUniqueness(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	t.Cleanup(func() { db.Exec("DELETE FROM usuarios") })

	ctx := context.Background()
	email := fmt.Sprintf("unique-%s@test.com", uuid.New().String())
	u1, _ := entity.NewUsuario(uuid.New().String(), "User 1", email, generateValidCPF(111111111))
	err := repo.Save(ctx, u1)
	if err != nil {
		t.Fatalf("First create failed: %v", err)
	}

	u2, _ := entity.NewUsuario(uuid.New().String(), "User 2", email, generateValidCPF(222222222))
	err = repo.Save(ctx, u2)
	if err == nil {
		t.Errorf("Expected error when creating user with duplicate email")
	}
}

func TestCPFUniqueness(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	t.Cleanup(func() { db.Exec("DELETE FROM usuarios") })

	ctx := context.Background()
	cpf := generateValidCPF(333333333)
	u1, _ := entity.NewUsuario(uuid.New().String(), "User 1", fmt.Sprintf("test1-%s@test.com", uuid.New().String()), cpf)
	err := repo.Save(ctx, u1)
	if err != nil {
		t.Fatalf("First create failed: %v", err)
	}

	u2, _ := entity.NewUsuario(uuid.New().String(), "User 2", fmt.Sprintf("test2-%s@test.com", uuid.New().String()), cpf)
	err = repo.Save(ctx, u2)
	if err == nil {
		t.Errorf("Expected error when creating user with duplicate CPF")
	}
}

func TestExistsByEmailAndCPF(t *testing.T) {
	db, repo := setupTestDB(t)
	if db == nil {
		return
	}
	t.Cleanup(func() { db.Exec("DELETE FROM usuarios") })

	ctx := context.Background()
	id := uuid.New().String()
	email := fmt.Sprintf("exists-%s@test.com", id)
	cpf := generateValidCPF(444444444)

	u, _ := entity.NewUsuario(id, "Test Exists", email, cpf)
	repo.Save(ctx, u)

	emailVO, _ := valueobject.NewEmail(email)
	existsEmail, err := repo.ExistsByEmail(ctx, emailVO)
	if err != nil {
		t.Errorf("ExistsByEmail failed: %v", err)
	}
	if !existsEmail {
		t.Errorf("Expected ExistsByEmail to be true")
	}

	cpfVO, _ := valueobject.NewCPF(cpf)
	existsCPF, err := repo.ExistsByCPF(ctx, cpfVO)
	if err != nil {
		t.Errorf("ExistsByCPF failed: %v", err)
	}
	if !existsCPF {
		t.Errorf("Expected ExistsByCPF to be true")
	}

	diffEmailVO, _ := valueobject.NewEmail("diff-" + email)
	existsDiff, _ := repo.ExistsByEmail(ctx, diffEmailVO)
	if existsDiff {
		t.Errorf("Expected ExistsByEmail for diff email to be false")
	}
}
