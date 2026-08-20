import { HttpInterceptorFn } from '@angular/common/http';
import { catchError, throwError } from 'rxjs';

export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  return next(req).pipe(
    catchError((err) => {
      console.log(err.status)
      if (err.status === 0) {
        console.error(
          `[HTTP] Falha de rede ou serviço indisponível: ${req.method} ${req.url}`,
        );
      }
      return throwError(() => err);
    }),
  );
};
