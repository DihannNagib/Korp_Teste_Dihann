import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import {
  provideHttpClientTesting,
  HttpTestingController,
} from '@angular/common/http/testing';

import { ProdutoService } from './produto.service';
import { environment } from '../../../environments/environment';

describe('ProdutoService', () => {
  let service: ProdutoService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        ProdutoService,
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    });
    service = TestBed.inject(ProdutoService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('deve listar produtos via GET', () => {
    service.listar().subscribe((produtos) => {
      expect(produtos.length).toBe(1);
      expect(produtos[0].codigo).toBe('P001');
    });

    const req = httpMock.expectOne(`${environment.apiUrlEstoque}/produtos`);
    expect(req.request.method).toBe('GET');
    req.flush([
      {
        id: 1,
        codigo: 'P001',
        descricao: 'Teste',
        saldo: 10,
        createdAt: '',
        updatedAt: '',
      },
    ]);
  });

  it('deve criar produto via POST', () => {
    const payload = { codigo: 'P002', descricao: 'Novo', saldo: 5 };

    service.criar(payload).subscribe((produto) => {
      expect(produto.codigo).toBe('P002');
    });

    const req = httpMock.expectOne(`${environment.apiUrlEstoque}/produtos`);
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual(payload);
    req.flush({ id: 2, ...payload, createdAt: '', updatedAt: '' });
  });
});
