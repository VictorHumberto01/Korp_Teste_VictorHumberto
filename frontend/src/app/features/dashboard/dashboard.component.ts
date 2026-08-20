import { Component, OnInit, OnDestroy, inject, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { RouterLink } from '@angular/router';
import { forkJoin, Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

import { UsuarioService } from '../usuarios/services/usuario.service';
import { ProdutoService } from '../produtos/services/produto.service';
import { NotaFiscalService } from '../notas-fiscais/services/nota-fiscal.service';

const LOW_STOCK_THRESHOLD = 5;

interface DashboardStats {
  totalUsuarios: number;
  totalProdutos: number;
  produtosComSaldoBaixo: number;
  totalNotas: number;
  notasAbertas: number;
  notasFechadas: number;
}

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, MatCardModule, MatIconModule, MatProgressSpinnerModule, RouterLink],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.scss'
})
export class DashboardComponent implements OnInit, OnDestroy {
  private usuarioService = inject(UsuarioService);
  private produtoService = inject(ProdutoService);
  private notaFiscalService = inject(NotaFiscalService);
  private destroy$ = new Subject<void>(); private cdr = inject(ChangeDetectorRef);

  isLoading = true;
  hasError = false;
  errorMessage = '';
  stats: DashboardStats | null = null;

  ngOnInit(): void {
    this.loadStats();
  }

  loadStats(): void {
    this.isLoading = true;
    this.hasError = false;

    // Os totais de usuários/notas vêm do total paginado da API; a contagem
    // de saldo baixo e a quebra aberta/fechada são calculadas sobre a
    // primeira página (até 100 itens) — suficiente para o escopo deste
    // painel, sem precisar de um endpoint de agregação dedicado.
    forkJoin({
      usuarios: this.usuarioService.getAll(1, 1),
      produtos: this.produtoService.getAll(1, 100),
      notas: this.notaFiscalService.getAll(1, 100)
    })
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: ({ usuarios, produtos, notas }) => {
          this.stats = {
            totalUsuarios: usuarios.total,
            totalProdutos: produtos.total,
            produtosComSaldoBaixo: (produtos.data || []).filter(p => p.saldo <= LOW_STOCK_THRESHOLD).length,
            totalNotas: notas.total,
            notasAbertas: (notas.data || []).filter(n => n.status === 'ABERTA').length,
            notasFechadas: (notas.data || []).filter(n => n.status === 'FECHADA').length
          };
          this.isLoading = false;
          this.cdr.detectChanges();
        },
        error: (err) => {
          this.errorMessage = err.message || 'Não foi possível carregar os indicadores. Verifique se os serviços estão no ar.';
          this.hasError = true;
          this.stats = null;
          this.isLoading = false;
          this.cdr.detectChanges();
        }
      });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
