# Active Context

## Current Phase
Hybrid Search, Re-ranking ve Tool geliştirme aşaması tamamlandı.

## Next Steps
- Streaming (Canlı yazı akışı) özelliğini Websocket tarafına eklemek.
- Memory katmanını Redis'e taşımak (Gelecek faz).

## Recent Changes
- `internal/vectorstore` içinde Hybrid Search ve Re-ranking (Keyword boost) mekanizması eklendi.
- `create_ticket` aracı (tool) Agent'a eklendi.
- Sistem promptu, otomatik arama yapacak ve duruma göre adım bekleyecek şekilde optimize edildi.
- `.gitignore` dosyası yapılandırıldı.
- `client/index.html` arayüzü, sağ altta açılabilen turuncu temalı "Teknik Asistan" widget'ına dönüştürüldü.
