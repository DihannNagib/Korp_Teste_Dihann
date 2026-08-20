export interface Produto { 
  id: number;
  codigo: string;
  descricao: string;
  saldo: number;
  createdAt: string;
  updateAt: string;
}

export interface CriarProdutoRequest { 
  codigo: string;
  descricao: string;
  saldo: number;
}