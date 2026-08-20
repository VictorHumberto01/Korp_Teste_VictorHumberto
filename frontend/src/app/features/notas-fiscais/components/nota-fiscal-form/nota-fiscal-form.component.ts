import { Component, OnInit, inject, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatDialogRef, MatDialogModule } from '@angular/material/dialog';
import { FormBuilder, FormArray, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { finalize } from 'rxjs/operators';

import { Produto } from '../../../produtos/models/produto.model';
import { ProdutoService } from '../../../produtos/services/produto.service';
import { NotaFiscalService } from '../../services/nota-fiscal.service';
import { NotificationService } from '../../../../core/services/notification.service';

@Component({
  selector: 'app-nota-fiscal-form',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule
  ],
  templateUrl: './nota-fiscal-form.component.html',
  styleUrl: './nota-fiscal-form.component.scss'
})
export class NotaFiscalFormComponent implements OnInit {
  private fb = inject(FormBuilder);
  private produtoService = inject(ProdutoService);
  private notaFiscalService = inject(NotaFiscalService);
  private notificationService = inject(NotificationService); private cdr = inject(ChangeDetectorRef);

  form!: FormGroup;
  produtos: Produto[] = [];
  isSubmitting = false;
  isLoadingProdutos = true;
  hasErrorProdutos = false;
  errorMessageProdutos = '';

  constructor(public dialogRef: MatDialogRef<NotaFiscalFormComponent>) {}

  ngOnInit(): void {
    this.form = this.fb.group({
      itens: this.fb.array([this.createItemGroup()])
    });

    this.loadProdutos();
  }

  loadProdutos(): void {
    this.isLoadingProdutos = true;
    this.hasErrorProdutos = false;

    this.produtoService.getAll(1, 100)
      .pipe(finalize(() => {
        this.isLoadingProdutos = false;
        this.cdr.detectChanges();
      }))
      .subscribe({
        next: (page) => {
          this.produtos = page.data || [];
          this.cdr.detectChanges();
        },
        error: (err) => {
          this.errorMessageProdutos = err.message || 'Erro ao carregar produtos.';
          this.hasErrorProdutos = true;
          this.produtos = [];
          this.notificationService.error(this.errorMessageProdutos);
          this.cdr.detectChanges();
        }
      });
  }

  get itens(): FormArray {
    return this.form.get('itens') as FormArray;
  }

  private createItemGroup(): FormGroup {
    return this.fb.group({
      produto_id: ['', Validators.required],
      quantidade: [1, [Validators.required, Validators.min(1)]]
    });
  }

  addItem(): void {
    this.itens.push(this.createItemGroup());
  }

  removeItem(index: number): void {
    if (this.itens.length > 1) {
      this.itens.removeAt(index);
    }
  }

  onSubmit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.isSubmitting = true;
    this.notaFiscalService.create(this.form.value.itens)
      .pipe(finalize(() => {
        this.isSubmitting = false;
        this.cdr.detectChanges();
      }))
      .subscribe({
        next: (res) => {
          this.notificationService.success(`Nota fiscal nº ${res.numero} criada com sucesso.`);
          this.dialogRef.close(res);
        },
        error: (err) => {
          this.notificationService.error(err.message || 'Erro ao criar nota fiscal.');
          this.cdr.detectChanges();
        }
      });
  }

  onCancel(): void {
    this.dialogRef.close();
  }
}
