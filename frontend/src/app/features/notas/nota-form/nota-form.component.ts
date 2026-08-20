import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import {
  FormArray,
  FormBuilder,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { finalize } from 'rxjs';

import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatAutocompleteModule } from '@angular/material/autocomplete';

import { NotaFiscalService } from '../../../core/services/nota-fiscal.service';
import { ProdutoService } from '../../../core/services/produto.service';

import { Produto } from '../../../core/models/produto.model';
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
    MatAutocompleteModule,
  ],
  templateUrl: './nota-form.component.html',
  styleUrl: './nota-form.component.scss',
})
export class NotaFormComponent implements OnInit {
  private fb = inject(FormBuilder);
  private notaService = inject(NotaFiscalService);
  private produtoService = inject(ProdutoService);
  private router = inject(Router);

  salvando = false;
  carregandoProdutos = false;
  erroGeral: string | null = null;

  produtos: Produto[] = [];

  form = this.fb.group({
    itens: this.fb.array([this.criarItem()]),
  });

  get itens(): FormArray {
    return this.form.get('itens') as FormArray;
  }

  ngOnInit(): void {
    this.carregarProdutos();
  }

  private carregarProdutos(): void {
    this.carregandoProdutos = true;
    this.erroGeral = null;
    this.produtoService
      .listar()
      .pipe(finalize(() => (this.carregandoProdutos = false)))
      .subscribe({
        next: (produtos) => {
          this.produtos = produtos;
        },
        error: () => {
          this.erroGeral =
            'Não foi possível carregar os produtos. O serviço de estoque pode estar indisponível.';
        },
      });
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
    if (this.itens.length > 1) {
      this.itens.removeAt(index);
    }
  }

  filtrarProdutos(valor: string): Produto[] {
    const termo = valor.toLowerCase().trim();
    if (!termo) {
      return this.produtos;
    }
    return this.produtos.filter(
      (produto) =>
        produto.codigo.toLowerCase().includes(termo) ||
        produto.descricao.toLowerCase().includes(termo),
    );
  }

  produtoExiste(codigo: string): boolean {
    return this.produtos.some((produto) => produto.codigo === codigo);
  }

  getProduto(codigo: string): Produto | undefined {
    return this.produtos.find((produto) => produto.codigo === codigo);
  }

  salvar(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    const itens = this.itens.getRawValue();
    const produtoDuplicado = this.encontrarProdutoDuplicado(itens);
    if (produtoDuplicado) {
      this.erroGeral = `O produto "${produtoDuplicado}" já foi adicionado à nota. Ajuste a quantidade no item existente.`;
      return;
    }
    const produtoInexistente = itens.find(
      (item) => !this.produtoExiste(item.produtoCodigo),
    );
    if (produtoInexistente) {
      this.erroGeral = `Produto "${produtoInexistente.produtoCodigo}" não encontrado. Selecione um produto existente.`;
      return;
    }
    const erroSaldo = this.validarSaldos(itens);
    if (erroSaldo) {
      this.erroGeral = erroSaldo;
      return;
    }
    this.salvando = true;
    this.erroGeral = null;
    this.notaService
      .criar({ itens })
      .pipe(finalize(() => (this.salvando = false)))
      .subscribe({
        next: (nota) => {
          this.router.navigate(['/notas', nota.numero]);
        },
        error: (err: HttpErrorResponse) => {
          this.erroGeral = this.extrairMensagem(err);
        },
      });
  }

  private extrairMensagem(err: HttpErrorResponse): string {
    const body = err.error as ApiErrorResponse;
    if (body?.erro) {
      return body.erro;
    }
    if (body?.erros?.length) {
      return body.erros.map((e) => e.erro).join(', ');
    }
    return 'Erro ao criar a nota fiscal. Tente novamente.';
  }

  private validarSaldos(
    itens: Array<{ produtoCodigo: string; quantidade: number }>,
  ): string | null {
    const quantidadesPorProduto = new Map<string, number>();
    for (const item of itens) {
      const quantidadeAtual =
        quantidadesPorProduto.get(item.produtoCodigo) ?? 0;
      quantidadesPorProduto.set(
        item.produtoCodigo,
        quantidadeAtual + item.quantidade,
      );
    }
    for (const [codigo, quantidadeTotal] of quantidadesPorProduto) {
      const produto = this.getProduto(codigo);
      if (!produto) {
        continue;
      }
      if (quantidadeTotal > produto.saldo) {
        return `O produto "${codigo}" possui saldo ${produto.saldo}, mas a nota utiliza ${quantidadeTotal} unidades.`;
      }
    }
    return null;
  }

  private encontrarProdutoDuplicado(
    itens: Array<{ produtoCodigo: string; quantidade: number }>,
  ): string | null {
    const codigos = new Set<string>();
    for (const item of itens) {
      if (codigos.has(item.produtoCodigo)) {
        return item.produtoCodigo;
      }
      codigos.add(item.produtoCodigo);
    }
    return null;
  }
}
