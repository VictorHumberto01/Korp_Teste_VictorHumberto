export interface Produto {
  id: string;
  codigo: string;
  descricao: string;
  saldo: number;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface ProdutoPage {
  data: Produto[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}
