export interface CampoErro { 
  campo: string;
  erro: string;
}

export interface ApiErrorResponse { 
  erro?: string;
  erros?: CampoErro[];
}