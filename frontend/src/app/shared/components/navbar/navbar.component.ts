import { Component } from '@angular/core';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { RouterLink, RouterLinkActive } from '@angular/router';

@Component({
  selector: 'app-navbar',
  standalone: true,
  imports: [MatToolbarModule, MatButtonModule, MatIconModule, MatTooltipModule, RouterLink, RouterLinkActive],
  template: `
    <mat-toolbar color="primary">
      <span>Sistema de Notas Fiscais</span>
      <span class="spacer"></span>
      <nav>
        <button mat-button routerLink="/usuarios" routerLinkActive="active-link">Usuários</button>
        <button mat-button routerLink="/produtos" routerLinkActive="active-link">Produtos</button>
        <span matTooltip="Em breve" matTooltipPosition="below">
          <button mat-button routerLink="/notas-fiscais" disabled>Notas Fiscais</button>
        </span>
      </nav>
    </mat-toolbar>
  `,
  styles: [`
    .spacer {
      flex: 1 1 auto;
    }
    .active-link {
      background: rgba(255, 255, 255, 0.1);
    }
    nav {
      display: flex;
      gap: 8px;
    }
  `]
})
export class NavbarComponent {}
