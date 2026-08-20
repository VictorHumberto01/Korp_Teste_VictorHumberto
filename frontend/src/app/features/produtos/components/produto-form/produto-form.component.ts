import { Component, Inject, OnInit, inject, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MAT_DIALOG_DATA, MatDialogRef, MatDialogModule } from '@angular/material/dialog';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatIconModule } from '@angular/material/icon';
import { finalize } from 'rxjs/operators';

import { Produto } from '../../models/produto.model';
import { ProdutoService } from '../../services/produto.service';
import { NotificationService } from '../../../../core/services/notification.service';

export interface ProdutoFormDialogData {
  produto?: Produto;
}

@Component({
  selector: 'app-produto-form',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatProgressSpinnerModule,
    MatProgressBarModule,
    MatIconModule
  ],
  templateUrl: './produto-form.component.html',
  styleUrl: './produto-form.component.scss'
})
export class ProdutoFormComponent implements OnInit {
  private fb = inject(FormBuilder);
  private produtoService = inject(ProdutoService);
  private notificationService = inject(NotificationService);
  private cdr = inject(ChangeDetectorRef);

  form!: FormGroup;
  isEditMode = false;
  isSubmitting = false;
  isSuggestingDesc = false;
  title = 'Novo Produto';

  constructor(
    public dialogRef: MatDialogRef<ProdutoFormComponent>,
    @Inject(MAT_DIALOG_DATA) public data: ProdutoFormDialogData
  ) {
    this.isEditMode = !!data?.produto;
    this.title = this.isEditMode ? 'Editar Produto' : 'Novo Produto';
  }

  ngOnInit(): void {
    this.initForm();
    if (this.isEditMode && this.data.produto) {
      this.form.patchValue({
        codigo: this.data.produto.codigo,
        descricao: this.data.produto.descricao,
        saldo: this.data.produto.saldo
      });
      this.form.get('codigo')?.disable();
      this.form.get('saldo')?.disable();
    }
  }

  private initForm(): void {
    this.form = this.fb.group({
      codigo: ['', [Validators.required, Validators.maxLength(50)]],
      descricao: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(200)]],
      saldo: [0, [Validators.required, Validators.min(0)]]
    });
  }

  suggestDescription(): void {
    const nomeBase = this.form.get('descricao')?.value;
    if (!nomeBase || nomeBase.length < 3) {
      this.notificationService.warning('Digite pelo menos 3 caracteres na descrição básica para gerar.');
      return;
    }

    this.isSuggestingDesc = true;
    this.produtoService.suggestDescription(nomeBase)
      .pipe(finalize(() => {
        this.isSuggestingDesc = false;
        this.cdr.detectChanges();
      }))
      .subscribe({
        next: (res) => {
          this.form.patchValue({ descricao: res.descricao });
          this.notificationService.success('Descrição gerada com IA!');
          this.cdr.detectChanges();
        },
        error: (err) => {
          this.notificationService.error('Erro ao gerar descrição com IA.');
          this.cdr.detectChanges();
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

    if (this.isEditMode && this.data.produto) {
      const updateData = {
        descricao: formValue.descricao,
        version: this.data.produto.version
      };

      this.produtoService.update(this.data.produto.id, updateData)
        .pipe(finalize(() => {
          this.isSubmitting = false;
          this.cdr.detectChanges();
        }))
        .subscribe({
          next: (res) => {
            this.notificationService.success('Produto atualizado com sucesso.');
            this.dialogRef.close(res);
          },
          error: (err) => {
            this.notificationService.error(err.message || 'Erro ao atualizar produto.');
            this.cdr.detectChanges();
          }
        });
    } else {
      const createData = {
        codigo: formValue.codigo,
        descricao: formValue.descricao,
        saldo: formValue.saldo
      };

      this.produtoService.create(createData)
        .pipe(finalize(() => {
          this.isSubmitting = false;
          this.cdr.detectChanges();
        }))
        .subscribe({
          next: (res) => {
            this.notificationService.success('Produto criado com sucesso.');
            this.dialogRef.close(res);
          },
          error: (err) => {
            this.notificationService.error(err.message || 'Erro ao criar produto.');
            this.cdr.detectChanges();
          }
        });
    }
  }

  onCancel(): void {
    this.dialogRef.close();
  }
}
