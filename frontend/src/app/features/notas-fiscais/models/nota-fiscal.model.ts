export type StatusNota = 'ABERTA' | 'FECHADA';

export interface ItemNota {
  produto_id: string;
  quantidade: number;
}

export interface NotaFiscal {
  id: string;
  numero: number;
  status: StatusNota;
  itens: ItemNota[];
  version: number;
  created_at: string;
  updated_at: string;
}

export interface NotaFiscalPage {
  data: NotaFiscal[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}
