import { HttpClient } from "@angular/common/http";
import { inject, Injectable } from "@angular/core";
import { environment } from "../../../environments/environment";
import { Observable } from "rxjs";
import { CriarProdutoRequest, Produto } from "../models/produto.model";

@Injectable({ providedIn: 'root' })
export class ProdutoService { 
  private http = inject(HttpClient);
  private baseUrl = `${environment.apiUrlEstoque}/produtos`;

  listar(): Observable<Produto[]> { 
    return this.http.get<Produto[]>(this.baseUrl);
  }

  criar(payload: CriarProdutoRequest): Observable<Produto> { 
    return this.http.post<Produto>(this.baseUrl, payload);
  }
}