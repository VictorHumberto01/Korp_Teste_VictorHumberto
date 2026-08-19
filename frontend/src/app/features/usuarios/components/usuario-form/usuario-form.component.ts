import { Component, Inject, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MAT_DIALOG_DATA, MatDialogRef, MatDialogModule } from '@angular/material/dialog';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { finalize } from 'rxjs/operators';

import { Usuario } from '../../models/usuario.model';
import { UsuarioService } from '../../services/usuario.service';
import { NotificationService } from '../../../../core/services/notification.service';

export interface UsuarioFormDialogData {
  usuario?: Usuario;
}

@Component({
  selector: 'app-usuario-form',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatProgressBarModule,
    MatIconModule,
    MatProgressSpinnerModule
  ],
  templateUrl: './usuario-form.component.html',
  styleUrl: './usuario-form.component.scss'
})
export class UsuarioFormComponent implements OnInit {
  private fb = inject(FormBuilder);
  private usuarioService = inject(UsuarioService);
  private notificationService = inject(NotificationService);

  form!: FormGroup;
  isEditMode = false;
  isSubmitting = false;
  isSuggestingBio = false;
  title = 'Novo Usuário';

  constructor(
    public dialogRef: MatDialogRef<UsuarioFormComponent>,
    @Inject(MAT_DIALOG_DATA) public data: UsuarioFormDialogData
  ) {
    this.isEditMode = !!data?.usuario;
    this.title = this.isEditMode ? 'Editar Usuário' : 'Novo Usuário';
  }

  ngOnInit(): void {
    this.initForm();
    if (this.isEditMode && this.data.usuario) {
      this.form.patchValue({
        nome: this.data.usuario.nome,
        email: this.data.usuario.email,
        cpf: this.data.usuario.cpf,
        bio: this.data.usuario.bio
      });
      this.form.get('cpf')?.disable();
    }
  }

  private initForm(): void {
    this.form = this.fb.group({
      nome: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(100)]],
      email: ['', [Validators.required, Validators.email]],
      cpf: ['', [Validators.required, Validators.pattern(/^\d{3}\.\d{3}\.\d{3}\-\d{2}$/)]],
      bio: ['']
    });
  }

  onCpfInput(event: Event): void {
    const input = event.target as HTMLInputElement;
    let value = input.value.replace(/\D/g, '');

    if (value.length > 11) {
      value = value.substring(0, 11);
    }

    let formattedValue = value;
    if (value.length > 9) {
      formattedValue = value.replace(/(\d{3})(\d{3})(\d{3})(\d{1,2})/, '$1.$2.$3-$4');
    } else if (value.length > 6) {
      formattedValue = value.replace(/(\d{3})(\d{3})(\d{1,3})/, '$1.$2.$3');
    } else if (value.length > 3) {
      formattedValue = value.replace(/(\d{3})(\d{1,3})/, '$1.$2');
    }

    input.value = formattedValue;
    this.form.get('cpf')?.setValue(formattedValue, { emitEvent: false });
  }

  suggestBio(): void {
    const nome = this.form.get('nome')?.value;
    const email = this.form.get('email')?.value;

    if (!nome || !email) {
      this.notificationService.warning('Preencha nome e email para sugerir a bio.');
      return;
    }

    this.isSuggestingBio = true;
    this.usuarioService.suggestBio(nome, email)
      .pipe(finalize(() => this.isSuggestingBio = false))
      .subscribe({
        next: (res) => {
          this.form.get('bio')?.setValue(res.bio);
        },
        error: () => {
          this.notificationService.error('IA indisponível no momento');
        }
      });
  }

  onSubmit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.isSubmitting = true;
    const formValue = this.form.getRawValue();
    const unformattedCpf = formValue.cpf.replace(/\D/g, '');

    if (this.isEditMode && this.data.usuario) {
      const updateData = {
        nome: formValue.nome,
        email: formValue.email,
        bio: formValue.bio,
        version: this.data.usuario.version
      };

      this.usuarioService.update(this.data.usuario.id, updateData)
        .pipe(finalize(() => this.isSubmitting = false))
        .subscribe({
          next: (res) => {
            this.notificationService.success('Usuário atualizado com sucesso.');
            this.dialogRef.close(res);
          },
          error: (err) => {
            this.notificationService.error(err.message || 'Erro ao atualizar usuário.');
          }
        });
    } else {
      const createData = {
        nome: formValue.nome,
        email: formValue.email,
        cpf: unformattedCpf
      };

      this.usuarioService.create(createData)
        .pipe(finalize(() => this.isSubmitting = false))
        .subscribe({
          next: (res) => {
            this.notificationService.success('Usuário criado com sucesso.');
            this.dialogRef.close(res);
          },
          error: (err) => {
            this.notificationService.error(err.message || 'Erro ao criar usuário.');
          }
        });
    }
  }

  onCancel(): void {
    this.dialogRef.close();
  }
}
