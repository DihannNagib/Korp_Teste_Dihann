import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatCardModule } from '@angular/material/card';
import { finalize } from 'rxjs';

import { NotaFiscalService } from '../../../core/services/nota-fiscal.service';
import { NotaFiscal } from '../../../core/models/nota-fiscal.model';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-nota-list',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatTableModule,
    MatButtonModule,
    MatChipsModule,
    MatIconModule,
    MatCardModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './nota-list.component.html',
  styleUrl: './nota-list.component.scss',
})
export class NotaListComponent implements OnInit {
  private notaService = inject(NotaFiscalService);

  notas: NotaFiscal[] = [];
  carregando = false;
  erro: string | null = null;
  colunas = ['numero', 'status', 'itens', 'acoes'];

  ngOnInit(): void {
    this.carregar();
    this.notaService.notaImpressaEvento$.subscribe((notaAtualizada) => {
      const idx = this.notas.findIndex(
        (n) => n.numero === notaAtualizada.numero,
      );
      if (idx >= 0) this.notas[idx] = notaAtualizada;
    });
  }

  carregar(): void {
    this.carregando = true;
    this.erro = null;

    this.notaService
      .listar()
      .pipe(finalize(() => (this.carregando = false)))
      .subscribe({
        next: (notas) => (this.notas = notas),
        error: () =>
          (this.erro = 'Não foi possível carregar as notas fiscais.'),
      });
  }
}
