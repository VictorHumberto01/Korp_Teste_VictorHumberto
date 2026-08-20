import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { BehaviorSubject, Observable, throwError } from 'rxjs';
import { catchError, retry, tap } from 'rxjs/operators';
import { environment } from '../../../../environments/environment';
import { Produto, ProdutoPage } from '../models/produto.model';

@Injectable({
  providedIn: 'root'
})
export class ProdutoService {
  private http = inject(HttpClient);
  private apiUrl = environment.estoqueApiUrl + '/produtos';

  private produtosSubject = new BehaviorSubject<Produto[]>([]);
  public produtos$ = this.produtosSubject.asObservable();

  private generateUUID(): string {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      const r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
    });
  }

  private handleError(error: any) {
    let errorMessage: string;

    if (error.status === 0) {
      // Falha de rede pura (conexão recusada, DNS, CORS bloqueado etc.) —
      // vem antes de tudo porque o objeto `error.error` nesse caso é um
      // ErrorEvent/ProgressEvent com uma mensagem crua do navegador (ex:
      // "NetworkError when attempting to fetch resource"), não algo pra
      // mostrar pro usuário.
      errorMessage = 'Não foi possível conectar ao serviço de estoque. Verifique se ele está no ar.';
    } else if (error.error?.error?.message) {
      // Envelope de erro do backend: { "error": { "code", "message" } }
      errorMessage = error.error.error.message;
    } else if (error.error?.message) {
      errorMessage = error.error.message;
    } else if (error.status === 404) {
      errorMessage = 'Produto não encontrado.';
    } else if (error.status === 502 || error.status === 503) {
      errorMessage = 'O serviço de estoque está temporariamente indisponível. Tente novamente em instantes.';
    } else if (error.status >= 500) {
      errorMessage = 'Ocorreu um erro no servidor. Tente novamente mais tarde.';
    } else if (error.status) {
      errorMessage = `Não foi possível completar a operação (código ${error.status}).`;
    } else {
      errorMessage = 'Ocorreu um erro desconhecido.';
    }

    return throwError(() => new Error(errorMessage));
  }

  getAll(page: number = 1, pageSize: number = 10): Observable<ProdutoPage> {
    return this.http.get<ProdutoPage>(`${this.apiUrl}?page=${page}&page_size=${pageSize}`).pipe(
      retry(2),
      tap(res => this.produtosSubject.next(res.data)),
      catchError(this.handleError)
    );
  }

  getById(id: string): Observable<Produto> {
    return this.http.get<Produto>(`${this.apiUrl}/${id}`).pipe(
      catchError(this.handleError)
    );
  }

  create(data: { codigo: string; descricao: string; saldo: number }): Observable<Produto> {
    const headers = new HttpHeaders({ 'Idempotency-Key': this.generateUUID() });
    return this.http.post<Produto>(this.apiUrl, data, { headers }).pipe(
      catchError(this.handleError)
    );
  }

  update(id: string, data: { descricao?: string; version: number }): Observable<Produto> {
    const headers = new HttpHeaders({ 'Idempotency-Key': this.generateUUID() });
    return this.http.put<Produto>(`${this.apiUrl}/${id}`, data, { headers }).pipe(
      catchError(this.handleError)
    );
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`).pipe(
      catchError(this.handleError)
    );
  }

  refreshList(page: number = 1, pageSize: number = 10): void {
    this.getAll(page, pageSize).subscribe();
  }

  suggestDescription(nome: string): Observable<{ descricao: string }> {
    return this.http.post<{ descricao: string }>(`${this.apiUrl}/suggest-description`, { nome }).pipe(
      catchError(this.handleError)
    );
  }
}
