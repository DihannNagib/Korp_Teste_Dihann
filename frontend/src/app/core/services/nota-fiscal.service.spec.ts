import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import {
  provideHttpClientTesting,
  HttpTestingController,
} from '@angular/common/http/testing';

import { NotaFiscalService } from './nota-fiscal.service';
import { environment } from '../../../environments/environment';
import { NotaFiscal } from '../models/nota-fiscal.model';

describe('NotaFiscalService', () => {
  let service: NotaFiscalService;
  let httpMock: HttpTestingController;

  const notaMock: NotaFiscal = {
    id: 1,
    numero: 1,
    status: 'ABERTA',
    itens: [{ id: 1, notaFiscalId: 1, produtoCodigo: 'P001', quantidade: 2 }],
    createdAt: '',
    updatedAt: '',
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        NotaFiscalService,
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    });
    service = TestBed.inject(NotaFiscalService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('deve criar nota via POST', () => {
    service
      .criar({ itens: [{ produtoCodigo: 'P001', quantidade: 2 }] })
      .subscribe((nota) => {
        expect(nota.status).toBe('ABERTA');
      });

    const req = httpMock.expectOne(`${environment.apiUrlFaturamento}/notas`);
    expect(req.request.method).toBe('POST');
    req.flush(notaMock);
  });

  it('deve imprimir nota e emitir evento notaImpressaEvento$', (done) => {
    const notaFechada = { ...notaMock, status: 'FECHADA' as const };

    service.notaImpressaEvento$.subscribe((nota) => {
      expect(nota.status).toBe('FECHADA');
      done();
    });

    service.imprimir(1).subscribe();

    const req = httpMock.expectOne(
      `${environment.apiUrlFaturamento}/notas/1/imprimir`,
    );
    expect(req.request.method).toBe('POST');
    req.flush(notaFechada);
  });
});
