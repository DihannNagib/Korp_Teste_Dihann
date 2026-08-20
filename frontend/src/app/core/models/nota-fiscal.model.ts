export interface ItemNotaFiscal { 
  id: number;
  notaFiscalId: number;
  produtoCodigo: string;
  quantidade: number;
}

export interface NotaFiscal { 
  id: number;
  numero: number;
  status: 'ABERTA' | 'FECHADA';
  itens: ItemNotaFiscal[];
  createdAt: string;
  updatedAt: string;
}

export interface ItemRequest { 
  produtoCodigo: string;
  quantidade: number;
}

export interface CriarNotaRequest { 
  itens: ItemRequest[];
}