import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { BehaviorSubject, Observable, throwError } from 'rxjs';
import { catchError, retry, tap } from 'rxjs/operators';
import { environment } from '../../../../environments/environment';
import { Usuario, UsuarioPage } from '../models/usuario.model';

@Injectable({
  providedIn: 'root'
})
export class UsuarioService {
  private http = inject(HttpClient);
  private apiUrl = environment.apiUrl + '/usuarios';

  private usuariosSubject = new BehaviorSubject<Usuario[]>([]);
  public usuarios$ = this.usuariosSubject.asObservable();

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
      errorMessage = 'Não foi possível conectar ao serviço de usuários. Verifique se ele está no ar.';
    } else if (error.error?.error?.message) {
      // Envelope de erro do backend: { "error": { "code", "message" } }
      errorMessage = error.error.error.message;
    } else if (error.error?.message) {
      errorMessage = error.error.message;
    } else if (error.status === 404) {
      errorMessage = 'Usuário não encontrado.';
    } else if (error.status === 502 || error.status === 503) {
      errorMessage = 'O serviço de usuários está temporariamente indisponível. Tente novamente em instantes.';
    } else if (error.status >= 500) {
      errorMessage = 'Ocorreu um erro no servidor. Tente novamente mais tarde.';
    } else if (error.status) {
      errorMessage = `Não foi possível completar a operação (código ${error.status}).`;
    } else {
      errorMessage = 'Ocorreu um erro desconhecido.';
    }

    return throwError(() => new Error(errorMessage));
  }

  getAll(page: number = 1, pageSize: number = 10): Observable<UsuarioPage> {
    return this.http.get<UsuarioPage>(`${this.apiUrl}?page=${page}&page_size=${pageSize}`).pipe(
      retry(2),
      tap(res => this.usuariosSubject.next(res.data)),
      catchError(this.handleError)
    );
  }

  getById(id: string): Observable<Usuario> {
    return this.http.get<Usuario>(`${this.apiUrl}/${id}`).pipe(
      catchError(this.handleError)
    );
  }

  create(data: { nome: string; email: string; cpf: string }): Observable<Usuario> {
    const headers = new HttpHeaders({ 'Idempotency-Key': this.generateUUID() });
    return this.http.post<Usuario>(this.apiUrl, data, { headers }).pipe(
      catchError(this.handleError)
    );
  }

  update(id: string, data: { nome?: string; email?: string; bio?: string; version: number }): Observable<Usuario> {
    const headers = new HttpHeaders({ 'Idempotency-Key': this.generateUUID() });
    return this.http.put<Usuario>(`${this.apiUrl}/${id}`, data, { headers }).pipe(
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
}
