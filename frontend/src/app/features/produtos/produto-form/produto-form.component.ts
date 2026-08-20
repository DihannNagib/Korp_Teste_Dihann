import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { HttpErrorResponse } from '@angular/common/http';
import { finalize } from 'rxjs';

import { ProdutoService } from '../../../core/services/produto.service';
import { ApiErrorResponse } from '../../../core/models/api-error.model';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-produto-form',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './produto-form.component.html',
  styleUrl: './produto-form.component.scss',
})
export class ProdutoFormComponent {
  private fb = inject(FormBuilder);
  private produtoService = inject(ProdutoService);
  private router = inject(Router);

  salvando = false;
  erroGeral: string | null = null;

  form = this.fb.nonNullable.group({
    codigo: ['', Validators.required],
    descricao: ['', Validators.required],
    saldo: [0, [Validators.required, Validators.min(0)]],
  });

  salvar(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.salvando = true;
    this.erroGeral = null;

    this.produtoService
      .criar(this.form.getRawValue())
      .pipe(finalize(() => (this.salvando = false)))
      .subscribe({
        next: () => this.router.navigate(['/produtos']),
        error: (err: HttpErrorResponse) =>
          (this.erroGeral = this.extrairMensagem(err)),
      });
  }

  private extrairMensagem(err: HttpErrorResponse): string {
    const body = err.error as ApiErrorResponse;
    if (body?.erro) return body.erro;
    if (body?.erros?.length) return body.erros.map((e) => e.erro).join(', ');
    return 'Erro ao salvar o produto. Tente novamente.';
  }
}