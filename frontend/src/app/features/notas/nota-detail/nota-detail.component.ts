import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatChipsModule } from '@angular/material/chips';
import { HttpErrorResponse } from '@angular/common/http';
import { finalize } from 'rxjs';

import { NotaFiscalService } from '../../../core/services/nota-fiscal.service';
import { NotaFiscal } from '../../../core/models/nota-fiscal.model';
import { ApiErrorResponse } from '../../../core/models/api-error.model';

@Component({
  selector: 'app-nota-detail',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatChipsModule,
  ],
  templateUrl: './nota-detail.component.html',
  styleUrl: './nota-detalhe.component.scss',
})
export class NotaDetailComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private notaService = inject(NotaFiscalService);
  private numeroAtual: number | null = null;

  nota: NotaFiscal | null = null;
  carregando = false;
  imprimindo = false;
  erro: string | null = null;

  ngOnInit(): void {
    this.numeroAtual = Number(this.route.snapshot.paramMap.get('numero'));
    if (this.numeroAtual) {
      this.carregar();
    } else {
      this.erro = 'Número da nota inválido.';
    }
  }

  carregar(): void {
    if (!this.numeroAtual) return;
    this.carregando = true;
    this.erro = null;

    this.notaService
      .buscarPorNumero(this.numeroAtual)
      .pipe(finalize(() => (this.carregando = false)))
      .subscribe({
        next: (nota) => (this.nota = nota),
        error: () => (this.erro = 'Nota fiscal não encontrada.'),
      });
  }

  imprimir(): void {
    if (!this.nota) return;

    this.imprimindo = true;
    this.erro = null;

    this.notaService
      .imprimir(this.nota.numero)
      .pipe(finalize(() => (this.imprimindo = false)))
      .subscribe({
        next: (notaAtualizada) => (this.nota = notaAtualizada),
        error: (err: HttpErrorResponse) =>
          (this.erro = this.extrairMensagem(err)),
      });
  }

  private extrairMensagem(err: HttpErrorResponse): string {
    const body = err.error as ApiErrorResponse;
    if (body?.erro) return body.erro;
    if (err.status === 0)
      return 'Não foi possível conectar ao servidor. Verifique sua conexão.';
    return 'Não foi possível imprimir a nota. Tente novamente.';
  }
}
