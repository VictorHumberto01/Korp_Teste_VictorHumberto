import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, retry } from 'rxjs/operators';
import { environment } from '../../../../environments/environment';
import { NotaFiscal, NotaFiscalPage, ItemNota } from '../models/nota-fiscal.model';

@Injectable({
  providedIn: 'root'
})
export class NotaFiscalService {
  private http = inject(HttpClient);
  private apiUrl = environment.faturamentoApiUrl + '/notas-fiscais';

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
      errorMessage = 'Não foi possível conectar ao serviço de notas fiscais. Verifique se ele está no ar.';
    } else if (error.error?.error?.message) {
      // Envelope de erro do backend: { "error": { "code", "message" } }
      errorMessage = error.error.error.message;
    } else if (error.error?.message) {
      errorMessage = error.error.message;
    } else if (error.status === 404) {
      errorMessage = 'Nota fiscal não encontrada.';
    } else if (error.status === 502 || error.status === 503) {
      errorMessage = 'O serviço de notas fiscais está temporariamente indisponível. Tente novamente em instantes.';
    } else if (error.status >= 500) {
      errorMessage = 'Ocorreu um erro no servidor. Tente novamente mais tarde.';
    } else if (error.status) {
      errorMessage = `Não foi possível completar a operação (código ${error.status}).`;
    } else {
      errorMessage = 'Ocorreu um erro desconhecido.';
    }

    return throwError(() => new Error(errorMessage));
  }

  getAll(page: number = 1, pageSize: number = 10): Observable<NotaFiscalPage> {
    return this.http.get<NotaFiscalPage>(`${this.apiUrl}?page=${page}&page_size=${pageSize}`).pipe(
      retry(2),
      catchError(this.handleError)
    );
  }

  getById(id: string): Observable<NotaFiscal> {
    return this.http.get<NotaFiscal>(`${this.apiUrl}/${id}`).pipe(
      catchError(this.handleError)
    );
  }

  create(itens: ItemNota[]): Observable<NotaFiscal> {
    const headers = new HttpHeaders({ 'Idempotency-Key': this.generateUUID() });
    return this.http.post<NotaFiscal>(this.apiUrl, { itens }, { headers }).pipe(
      catchError(this.handleError)
    );
  }

  imprimir(id: string): Observable<NotaFiscal> {
    const headers = new HttpHeaders({ 'Idempotency-Key': this.generateUUID() });
    return this.http.post<NotaFiscal>(`${this.apiUrl}/${id}/imprimir`, {}, { headers }).pipe(
      catchError(this.handleError)
    );
  }
}
