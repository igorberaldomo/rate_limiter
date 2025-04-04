Objetivo: Desenvolver um rate limiter em Go que possa ser configurado para limitar o número máximo de requisições por segundo com base em um endereço IP específico ou em um token de acesso.

    O código disponibiliza um servidor no http://localhost:8080/
        caso não tenha sido disponibilizado um API_KEY ele vai alocar 5 requests por 10 segundos
        depois desses 5 requests ele vai blockear o acesso
    Api_key
        você vai enviar para um servidor http://localhost:8080/?API_KEY=apikey
        isso vai alocar 10 requests a cada 10 segundos depois disso vai bloquear o acesso

    para subir o código levante o docker composer
    para mudar os valores de qualquer variavel modifique o .env
    todos os requests são gravados de forma temporária no redis
    para fazer requisições em massa utilize go run test/test.go
    
