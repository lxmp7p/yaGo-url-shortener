В коде проекта были слишком затратные операции взаимодействия с файловым хранилищем.
Для хранения, изменения и получения данных из файлового хранилища файл открывался при каждом запросе.
Т.е каждый перед каждым запросом выполнялась операция ОТКРЫТИЯ файла, а после выполнения файл закрывался.

С помощью pprof удалось выявить узкое место и оперативно отрефакторить затратные функции.
Благодаря этому, удалось снизить потребление ЦПУ и памяти во время работы сервиса.
-> % go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
File: __debug_bin2595923181
Build ID: 3bb1f32e7de4bb791164d0e401f84caee9560216
Type: inuse_space
Time: 2026-06-02 11:50:57 MSK
Showing nodes accounting for -2156.17kB, 50.68% of 4254.75kB total
      flat  flat%   sum%        cum   cum%
   -1123kB 26.39% 26.39%    -1123kB 26.39%  github.com/lxmp7p/yaGo-url-shortener/internal/repository.(*Cache).Save
 -521.05kB 12.25% 38.64%  -521.05kB 12.25%  github.com/lib/pq.map.init.0
 -512.12kB 12.04% 50.68%  -512.12kB 12.04%  net/http.ListenAndServe
  512.05kB 12.03% 38.64%   512.05kB 12.03%  context.(*cancelCtx).Done
 -512.05kB 12.03% 50.68%  -512.05kB 12.03%  runtime.mallocgc
         0     0% 50.68%   512.05kB 12.03%  database/sql.(*DB).connectionOpener
         0     0% 50.68%    -1123kB 26.39%  github.com/go-chi/chi/v5.(*Mux).Mount.func1
         0     0% 50.68%    -1123kB 26.39%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 50.68%    -1123kB 26.39%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 50.68%  -521.05kB 12.25%  github.com/lib/pq.init
         0     0% 50.68%    -1123kB 26.39%  github.com/lxmp7p/yaGo-url-shortener/internal/handler.(*Handler).AuthMiddleware.func1
         0     0% 50.68%    -1123kB 26.39%  github.com/lxmp7p/yaGo-url-shortener/internal/handler.CompressMiddleware.func1.1
         0     0% 50.68%    -1123kB 26.39%  github.com/lxmp7p/yaGo-url-shortener/internal/handler.LoggingMiddleware.func1.1
         0     0% 50.68%    -1123kB 26.39%  github.com/lxmp7p/yaGo-url-shortener/internal/service.(*ShortenerService).CreateShortURL
         0     0% 50.68%  -512.12kB 12.04%  main.main.func1
         0     0% 50.68%    -1123kB 26.39%  net/http.(*conn).serve
         0     0% 50.68%    -1123kB 26.39%  net/http.HandlerFunc.ServeHTTP
         0     0% 50.68%    -1123kB 26.39%  net/http.serverHandler.ServeHTTP
         0     0% 50.68%  -512.05kB 12.03%  runtime.acquireSudog
         0     0% 50.68%  -521.05kB 12.25%  runtime.doInit
         0     0% 50.68%  -521.05kB 12.25%  runtime.doInit1
         0     0% 50.68%  -512.05kB 12.03%  runtime.gcBgMarkWorker
         0     0% 50.68%  -512.05kB 12.03%  runtime.gcMarkDone
         0     0% 50.68%  -521.05kB 12.25%  runtime.main
         0     0% 50.68%  -512.05kB 12.03%  runtime.newobject
         0     0% 50.68%  -512.05kB 12.03%  runtime.semacquire
         0     0% 50.68%  -512.05kB 12.03%  runtime.semacquire1