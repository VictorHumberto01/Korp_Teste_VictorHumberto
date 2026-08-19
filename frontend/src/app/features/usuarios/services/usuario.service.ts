import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { BehaviorSubject, Observable, throwError } from 'rxjs';
import { catchError, retry, tap } from 'rxjs/operators';
import { environment } from '../../../../environments/environment';
import { Usuario, UsuarioPage, SuggestBioResponse } from '../models/usuario.model';

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
    let errorMessage = 'Ocorreu um erro desconhecido.';
    if (error.error instanceof ErrorEvent) {
      errorMessage = `Erro: ${error.error.message}`;
    } else if (error.status) {
      if (error.error && error.error.message) {
        errorMessage = error.error.message;
      } else {
        errorMessage = `Código do erro: ${error.status}, mensagem: ${error.message}`;
      }
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

  suggestBio(nome: string, email: string): Observable<SuggestBioResponse> {
    return this.http.post<SuggestBioResponse>(`${this.apiUrl}/suggest-bio`, { nome, email }).pipe(
      catchError(this.handleError)
    );
  }

  refreshList(page: number = 1, pageSize: number = 10): void {
    this.getAll(page, pageSize).subscribe();
  }
}
