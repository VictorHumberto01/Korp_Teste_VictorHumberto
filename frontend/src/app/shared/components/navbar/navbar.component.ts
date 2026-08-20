import { Component } from '@angular/core';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink, RouterLinkActive } from '@angular/router';

@Component({
  selector: 'app-navbar',
  standalone: true,
  imports: [MatToolbarModule, MatButtonModule, MatIconModule, RouterLink, RouterLinkActive],
  template: `
    <mat-toolbar color="primary">
      <a routerLink="/dashboard" class="brand">
        <mat-icon>receipt_long</mat-icon>
        <span>Sistema de Notas Fiscais</span>
      </a>
      <span class="spacer"></span>
      <nav>
        <button mat-button routerLink="/dashboard" routerLinkActive="active-link">
          <mat-icon>dashboard</mat-icon>
          Visão Geral
        </button>
        <button mat-button routerLink="/usuarios" routerLinkActive="active-link">
          <mat-icon>group</mat-icon>
          Usuários
        </button>
        <button mat-button routerLink="/produtos" routerLinkActive="active-link">
          <mat-icon>inventory_2</mat-icon>
          Produtos
        </button>
        <button mat-button routerLink="/notas-fiscais" routerLinkActive="active-link">
          <mat-icon>receipt_long</mat-icon>
          Notas Fiscais
        </button>
      </nav>
    </mat-toolbar>
  `,
  styles: [`
    .brand {
      display: flex;
      align-items: center;
      gap: 8px;
      color: inherit;
      text-decoration: none;
      font-size: 20px;
      font-weight: 500;
    }
    .spacer {
      flex: 1 1 auto;
    }
    .active-link {
      background: rgba(255, 255, 255, 0.14);
    }
    nav {
      display: flex;
      gap: 4px;
    }
    nav button {
      display: flex;
      align-items: center;
      gap: 6px;
    }
    nav mat-icon {
      font-size: 20px;
      width: 20px;
      height: 20px;
    }
  `]
})
export class NavbarComponent {}
