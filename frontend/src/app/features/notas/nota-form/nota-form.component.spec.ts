import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import {
  provideHttpClientTesting,
  HttpTestingController,
} from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';
import { Router } from '@angular/router';

import { NotaFormComponent } from './nota-form.component';
import { environment } from '../../../../environments/environment';

describe('NotaFormComponent', () => {
  let fixture: ComponentFixture<NotaFormComponent>;
  let component: NotaFormComponent;
  let httpMock: HttpTestingController;
  let router: Router;

  const produtosMock = [
    {
      id: 1,
      codigo: 'P001',
      descricao: 'Produto teste',
      saldo: 10,
    },
    {
      id: 2,
      codigo: 'P002',
      descricao: 'Outro produto',
      saldo: 5,
    },
  ];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NotaFormComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(NotaFormComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
    router = TestBed.inject(Router);
    fixture.detectChanges();
    const req = httpMock.expectOne(`${environment.apiUrlEstoque}/produtos`);
    expect(req.request.method).toBe('GET');
    req.flush(produtosMock);
  });

  afterEach(() => httpMock.verify());

  it('deve iniciar com um item no FormArray', () => {
    expect(component.itens.length).toBe(1);
  });

  it('deve adicionar e remover itens', () => {
    component.adicionarItem();
    expect(component.itens.length).toBe(2);
    component.removerItem(1);
    expect(component.itens.length).toBe(1);
  });

  it('nao deve remover o ultimo item restante', () => {
    component.removerItem(0);
    expect(component.itens.length).toBe(1);
  });

  it('nao deve submeter formulario invalido', () => {
    component.itens.at(0).patchValue({
      produtoCodigo: '',
      quantidade: 1,
    });
    component.salvar();
    httpMock.expectNone(`${environment.apiUrlFaturamento}/notas`);
  });

  it('nao deve criar nota com produto inexistente', () => {
    component.itens.at(0).patchValue({
      produtoCodigo: 'P999',
      quantidade: 2,
    });
    component.salvar();
    expect(component.erroGeral).toBe(
      'Produto "P999" não encontrado. Selecione um produto existente.',
    );
    httpMock.expectNone(`${environment.apiUrlFaturamento}/notas`);
  });

  it('nao deve permitir produto duplicado na nota', () => {
    component.itens.at(0).patchValue({
      produtoCodigo: 'P001',
      quantidade: 2,
    });
    component.adicionarItem();
    component.itens.at(1).patchValue({
      produtoCodigo: 'P001',
      quantidade: 3,
    });

    component.salvar();
    expect(component.erroGeral).toBe(
      'O produto "P001" já foi adicionado à nota. Ajuste a quantidade no item existente.',
    );
    httpMock.expectNone(`${environment.apiUrlFaturamento}/notas`);
  });

  it('nao deve criar nota quando a quantidade excede o saldo', () => {
    component.itens.at(0).patchValue({
      produtoCodigo: 'P001',
      quantidade: 11,
    });
    component.salvar();
    expect(component.erroGeral).toBe(
      'O produto "P001" possui saldo 10, mas a nota utiliza 11 unidades.',
    );
    httpMock.expectNone(`${environment.apiUrlFaturamento}/notas`);
  });

  it('nao deve criar nota quando a quantidade excede o saldo de um produto', () => {
    component.itens.at(0).patchValue({
      produtoCodigo: 'P002',
      quantidade: 6,
    });
    component.salvar();
    expect(component.erroGeral).toBe(
      'O produto "P002" possui saldo 5, mas a nota utiliza 6 unidades.',
    );
    httpMock.expectNone(`${environment.apiUrlFaturamento}/notas`);
  });

  it('deve criar nota e navegar para o detalhe', () => {
    const navigateSpy = spyOn(router, 'navigate');
    component.itens.at(0).patchValue({
      produtoCodigo: 'P001',
      quantidade: 2,
    });
    component.salvar();
    const req = httpMock.expectOne(`${environment.apiUrlFaturamento}/notas`);
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({
      itens: [
        {
          produtoCodigo: 'P001',
          quantidade: 2,
        },
      ],
    });
    req.flush({
      id: 1,
      numero: 5,
      status: 'ABERTA',
      itens: [
        {
          id: 1,
          notaFiscalId: 1,
          produtoCodigo: 'P001',
          quantidade: 2,
        },
      ],
      createdAt: '',
      updatedAt: '',
    });
    expect(navigateSpy).toHaveBeenCalledWith(['/notas', 5]);
  });
});
