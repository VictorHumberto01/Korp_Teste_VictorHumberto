import { Routes } from '@angular/router';

export const USUARIOS_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./components/usuario-list/usuario-list.component').then(m => m.UsuarioListComponent)
  }
];
