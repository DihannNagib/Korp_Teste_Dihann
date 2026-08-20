import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { finalize } from 'rxjs';

import { ProdutoService } from '../../../core/services/produto.service';
import { Produto } from '../../../core/models/produto.model';

@Component({
  selector: 'app-produto-list',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './produto-list.component.html',
})
export class ProdutoListComponent implements OnInit {
  private produtoService = inject(ProdutoService);

  produtos: Produto[] = [];
  carregando = false;
  erro: string | null = null;
  colunas = ['codigo', 'descricao', 'saldo'];

  ngOnInit(): void {
    this.carregar();
  }

  carregar(): void {
    this.carregando = true;
    this.erro = null;

    this.produtoService
      .listar()
      .pipe(finalize(() => (this.carregando = false)))
      .subscribe({
        next: (produtos) => (this.produtos = produtos),
        error: () =>
          (this.erro =
            'Não foi possível carregar os produtos. O serviço de estoque pode estar indisponível.'),
      });
  }
}
