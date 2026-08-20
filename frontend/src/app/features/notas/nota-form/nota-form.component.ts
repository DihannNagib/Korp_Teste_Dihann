import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import {
  FormArray,
  FormBuilder,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { HttpErrorResponse } from '@angular/common/http';
import { finalize } from 'rxjs';

import { NotaFiscalService } from '../../../core/services/nota-fiscal.service';
import { ApiErrorResponse } from '../../../core/models/api-error.model';

@Component({
  selector: 'app-nota-form',
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
  templateUrl: './nota-form.component.html',
})
export class NotaFormComponent {
  private fb = inject(FormBuilder);
  private notaService = inject(NotaFiscalService);
  private router = inject(Router);

  salvando = false;
  erroGeral: string | null = null;

  form = this.fb.group({
    itens: this.fb.array([this.criarItem()]),
  });

  get itens(): FormArray {
    return this.form.get('itens') as FormArray;
  }

  private criarItem() {
    return this.fb.nonNullable.group({
      produtoCodigo: ['', Validators.required],
      quantidade: [1, [Validators.required, Validators.min(1)]],
    });
  }

  adicionarItem(): void {
    this.itens.push(this.criarItem());
  }

  removerItem(index: number): void {
    if (this.itens.length > 1) this.itens.removeAt(index);
  }

  salvar(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.salvando = true;
    this.erroGeral = null;

    const payload = { itens: this.itens.getRawValue() };

    this.notaService
      .criar(payload)
      .pipe(finalize(() => (this.salvando = false)))
      .subscribe({
        next: (nota) => this.router.navigate(['/notas', nota.numero]),
        error: (err: HttpErrorResponse) =>
          (this.erroGeral = this.extrairMensagem(err)),
      });
  }

  private extrairMensagem(err: HttpErrorResponse): string {
    const body = err.error as ApiErrorResponse;
    if (body?.erro) return body.erro;
    if (body?.erros?.length) return body.erros.map((e) => e.erro).join(', ');
    return 'Erro ao criar a nota fiscal. Tente novamente.';
  }
}
