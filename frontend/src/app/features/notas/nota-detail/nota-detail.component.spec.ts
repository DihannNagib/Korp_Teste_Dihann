import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import {
  provideHttpClientTesting,
  HttpTestingController,
} from '@angular/common/http/testing';
import { ActivatedRoute, convertToParamMap } from '@angular/router';

import { NotaDetailComponent } from './nota-detail.component';
import { environment } from '../../../../environments/environment';

describe('NotaDetailComponent', () => {
  let fixture: ComponentFixture<NotaDetailComponent>;
  let component: NotaDetailComponent;
  let httpMock: HttpTestingController;

  const notaBase = {
    id: 1,
    numero: 7,
    status: 'ABERTA' as const,
    itens: [{ id: 1, notaFiscalId: 1, produtoCodigo: 'P001', quantidade: 2 }],
    createdAt: '',
    updatedAt: '',
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NotaDetailComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: { paramMap: convertToParamMap({ numero: '7' }) },
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(NotaDetailComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();

    httpMock
      .expectOne(`${environment.apiUrlFaturamento}/notas/7`)
      .flush(notaBase);
  });

  afterEach(() => httpMock.verify());

  it('deve carregar a nota ao iniciar', () => {
    expect(component.nota?.numero).toBe(7);
    expect(component.nota?.status).toBe('ABERTA');
  });

  it('impressao com sucesso deve atualizar status para FECHADA', () => {
    component.imprimir();
    expect(component.imprimindo).toBeTrue();

    const req = httpMock.expectOne(
      `${environment.apiUrlFaturamento}/notas/7/imprimir`,
    );
    req.flush({ ...notaBase, status: 'FECHADA' });

    expect(component.imprimindo).toBeFalse();
    expect(component.nota?.status).toBe('FECHADA');
    expect(component.erro).toBeNull();
  });

  it('falha na impressao (estoque indisponivel) deve manter nota ABERTA e mostrar erro', () => {
    component.imprimir();

    const req = httpMock.expectOne(
      `${environment.apiUrlFaturamento}/notas/7/imprimir`,
    );
    req.flush(
      { erro: 'falha ao comunicar com o servico de estoque' },
      { status: 503, statusText: 'Service Unavailable' },
    );

    expect(component.imprimindo).toBeFalse();
    expect(component.nota?.status).toBe('ABERTA');
    expect(component.erro).toBe('falha ao comunicar com o servico de estoque');
  });
});
