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
    component.itens.at(0).patchValue({ produtoCodigo: '', quantidade: 1 });
    component.salvar();
    httpMock.expectNone(`${environment.apiUrlFaturamento}/notas`);
  });

  it('deve criar nota e navegar para o detalhe', () => {
    const navigateSpy = spyOn(router, 'navigate');

    component.itens.at(0).patchValue({ produtoCodigo: 'P001', quantidade: 2 });
    component.salvar();

    const req = httpMock.expectOne(`${environment.apiUrlFaturamento}/notas`);
    req.flush({
      id: 1,
      numero: 5,
      status: 'ABERTA',
      itens: [{ id: 1, notaFiscalId: 1, produtoCodigo: 'P001', quantidade: 2 }],
      createdAt: '',
      updatedAt: '',
    });

    expect(navigateSpy).toHaveBeenCalledWith(['/notas', 5]);
  });
});
