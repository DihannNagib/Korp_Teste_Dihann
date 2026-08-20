import { HttpClient } from "@angular/common/http";
import { inject, Injectable } from "@angular/core";
import { environment } from "../../../environments/environment";
import { Observable, Subject, tap } from "rxjs";
import { CriarNotaRequest, NotaFiscal } from "../models/nota-fiscal.model";

@Injectable({ providedIn: 'root' })
export class NotaFiscalService {
  private http = inject(HttpClient);
  private baseUrl = `${environment.apiUrlFaturamento}/notas`;

  private notaImpressa$ = new Subject<NotaFiscal>();
  readonly notaImpressaEvento$ = this.notaImpressa$.asObservable();

  listar(): Observable<NotaFiscal[]> {
    return this.http.get<NotaFiscal[]>(this.baseUrl);
  }

  buscarPorNumero(numero: number): Observable<NotaFiscal> {
    return this.http.get<NotaFiscal>(`${this.baseUrl}/${numero}`);
  }

  criar(payload: CriarNotaRequest): Observable<NotaFiscal> {
    return this.http.post<NotaFiscal>(this.baseUrl, payload);
  }

  imprimir(numero: number): Observable<NotaFiscal> {
    return this.http
      .post<NotaFiscal>(`${this.baseUrl}/${numero}/imprimir`, {})
      .pipe(tap((nota) => this.notaImpressa$.next(nota)));
  }
}