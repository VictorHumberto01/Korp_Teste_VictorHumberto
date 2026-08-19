import { Routes } from '@angular/router';

export const PRODUTOS_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./components/produto-list/produto-list.component').then(m => m.ProdutoListComponent)
  }
];
