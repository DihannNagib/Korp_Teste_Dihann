import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '**', redirectTo: 'produtos', pathMatch: 'full' },
  {
    path: 'produtos',
    loadComponent: () =>
      import('./features/produtos/produto-list/produto-list.component').then(
        (m) => m.ProdutoListComponent,
      ),
  },
];
