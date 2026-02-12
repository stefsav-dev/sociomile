# Sociomile - Customer Support Platform

Sociomile adalah platform customer support yang memungkinkan komunikasi real-time antara customer dan agent. Sistem ini dibangun dengan Go Fiber untuk backend dan Next.js untuk frontend, menggunakan MySQL sebagai database utama dan Redis untuk caching.

- Backend :

    - Golang
    - Fiber 
    - MySQL
    - Redis 
    - Gorm
    - JWT
    - godotnev


- Frontend : 
    - NextJS
    - ShadcnUI
    - TailwindCSS
    - Axios

untuk menjalankan program tersebut :
 
 Backend :

 untuk backend nanti dengan cara pada cmd ketik docker compose up nanti pada program backend langsung bisa di jalankan dengan url "http://localhost:8000"

 Frontend :

 untuk frontend jalankan program dengan ketik pada folder nya npm run dev maka sudah tersambung dengan backend nya


 untuk daftar endpoint nya :

    Auth :
     - http://127.0.0.1:8000/api/auth/login
     - http://127.0.0.1:8000/api/auth/register
     - http://127.0.0.1:8000/api/auth/logout

     User : 
     - http://127.0.0.1:8000/api/user/profile
     - http://127.0.0.1:8000/api/user/channels/1/messages
     - http://127.0.0.1:8000/api/user/channels

     Agent :
     - http://127.0.0.1:8000/api/agent/conversations?status=all&limit=10&offset=0
     - http://127.0.0.1:8000/api/agent/channels/available
     - http://127.0.0.1:8000/api/agent/channels/stats
     - http://127.0.0.1:8000/api/agent/channels/1/assign
     - http://127.0.0.1:8000/api/agent/channels/1/messages
     - http://127.0.0.1:8000/api/agent/channels/1/close

    
    untuk program ini di bagian backend nya sudah semua untuk service-service nya dan endpoint nya
    namun di bagian frontend pada chat nya masih mengalami bug belum bisa mengirim dari user ke agent secara realtime
    dan frontend hanya bisa di bagian auth login multirole dan logout.
    dikarenakan saya masih bekerja bug/error dan fitur chat untuk di bagian frontend nya belum selesai. saya minta maaf.

    Demikian hasil pengerjaan saya. Terima Kasih

