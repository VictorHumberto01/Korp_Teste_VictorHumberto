import { Component, OnInit, OnDestroy, AfterViewInit, ViewChild, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule, MatTableDataSource } from '@angular/material/table';
import { MatPaginator, MatPaginatorModule } from '@angular/material/paginator';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { Subject } from 'rxjs';
import { takeUntil, finalize } from 'rxjs/operators';

import { NotaFiscalService } from '../../services/nota-fiscal.service';
import { NotificationService } from '../../../../core/services/notification.service';
import { NotaFiscal } from '../../models/nota-fiscal.model';
import { NotaFiscalFormComponent } from '../nota-fiscal-form/nota-fiscal-form.component';

@Component({
  selector: 'app-nota-fiscal-list',
  standalone: true,
  imports: [
    CommonModule,
    MatTableModule,
    MatPaginatorModule,
    MatSortModule,
    MatButtonModule,
    MatIconModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    MatTooltipModule,
    MatDialogModule
  ],
  templateUrl: './nota-fiscal-list.component.html',
  styleUrl: './nota-fiscal-list.component.scss'
})
export class NotaFiscalListComponent implements OnInit, OnDestroy, AfterViewInit {
  private notaFiscalService = inject(NotaFiscalService);
  private notificationService = inject(NotificationService);
  private dialog = inject(MatDialog);
  private destroy$ = new Subject<void>();

  displayedColumns: string[] = ['numero', 'status', 'itens', 'acoes'];
  dataSource = new MatTableDataSource<NotaFiscal>([]);
  isLoading = true;
  hasError = false;
  errorMessage = '';
  totalRecords = 0;
  imprimindoIds = new Set<string>();

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  ngOnInit(): void {
    this.loadData();
  }

  ngAfterViewInit(): void {
    this.dataSource.paginator = this.paginator;
    this.dataSource.sort = this.sort;

    this.paginator.page.pipe(takeUntil(this.destroy$)).subscribe(() => {
      this.loadData();
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  loadData(): void {
    this.isLoading = true;
    this.hasError = false;
    const pageIndex = this.paginator ? this.paginator.pageIndex + 1 : 1;
    const pageSize = this.paginator ? this.paginator.pageSize : 10;

    this.notaFiscalService.getAll(pageIndex, pageSize)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (page) => {
          this.dataSource.data = page.data || [];
          this.totalRecords = page.total;
          this.isLoading = false;
        },
        error: (err) => {
          this.errorMessage = err.message || 'Erro ao carregar notas fiscais.';
          this.hasError = true;
          this.dataSource.data = [];
          this.totalRecords = 0;
          this.isLoading = false;
          this.notificationService.error(this.errorMessage);
        }
      });
  }

  openFormDialog(): void {
    const dialogRef = this.dialog.open(NotaFiscalFormComponent, {
      width: '600px',
      data: {}
    });

    dialogRef.afterClosed().pipe(takeUntil(this.destroy$)).subscribe(result => {
      if (result) {
        this.loadData();
      }
    });
  }

  imprimir(nota: NotaFiscal): void {
    if (nota.status !== 'ABERTA' || this.imprimindoIds.has(nota.id)) return;

    this.imprimindoIds.add(nota.id);
    this.notaFiscalService.imprimir(nota.id)
      .pipe(
        takeUntil(this.destroy$),
        finalize(() => this.imprimindoIds.delete(nota.id))
      )
      .subscribe({
        next: () => {
          this.notificationService.success(`Nota fiscal nº ${nota.numero} impressa com sucesso.`);
          this.loadData();
        },
        error: (err) => {
          this.notificationService.error(err.message || 'Erro ao imprimir nota fiscal.');
        }
      });
  }

  isImprimindo(id: string): boolean {
    return this.imprimindoIds.has(id);
  }
}
