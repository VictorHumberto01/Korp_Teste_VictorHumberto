import { Routes } from '@angular/router';

export const NOTAS_FISCAIS_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./components/nota-fiscal-list/nota-fiscal-list.component').then(m => m.NotaFiscalListComponent)
  }
];
