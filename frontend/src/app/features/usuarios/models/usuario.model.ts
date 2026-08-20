export interface Usuario {
  id: string;
  nome: string;
  email: string;
  cpf: string;
  bio: string;
  ativo: boolean;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface UsuarioPage {
  data: Usuario[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}
