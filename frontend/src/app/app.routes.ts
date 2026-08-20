import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', redirectTo: 'produtos', pathMatch: 'full' },
  {
    path: 'produtos',
    loadComponent: () =>
      import('./features/produtos/produto-list/produto-list.component').then(
        (m) => m.ProdutoListComponent,
      ),
  },
  {
    path: 'produtos/novo',
    loadComponent: () =>
      import('./features/produtos/produto-form/produto-form.component').then(
        (m) => m.ProdutoFormComponent,
      ),
  },
  {
    path: '**',
    redirectTo: 'produtos',
  },
];
