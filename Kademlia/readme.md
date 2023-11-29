- dar olhada docker-compose. Note que o campo networks, qdo rodar o build e up, vai criar 
uma interface com subrede definida na linha 43 e os nós vão sendo adicionados com endereço fixo
nesta subrede. Estes nós nao significam que eles já estão no CHORD, apenas indicam nós ativos. Ideia: ter dois arquivos Dockerfile-client (1 e 2), um deles vai fazer com que aqueles nós
já estejam na rede CHORD, e o outro vai criar nós que nós vamos colocar na rede chord de modo
manual, fazendo comando no terminal desse cara dizendo pra ele entrar (no caso de definir a 
função que faz ele entrar)

- os arquivos Dockerfile-client seguem o mesmo padrão.

- pra rodar, abrir terminal no local do arquivo docker-compose
docker compose build
docker compose up

- ctrl C e docker compose down pra terminar.

- para ver containers em outro terminal:
docker attach ci

- erro de network: ocorre quando nao ha interface no ifconfig dos ips configurados no docker-compose. apos comando abaixo, build e up novamente.
docker network prune

- como fazer no docker-compose para abrir terminal ao inves de processo ja rodando:
no services que voce quiser, vai ter o campo stdin_open e tty iguais a true, como abaixo.
E também precisa tirar o campo CMD do Dockerfile2:

container2:
    build:
      context: .
      dockerfile: "Dockerfile2"
    networks:
      mynetwork:
        ipv4_address: 172.18.0.3
    stdin_open: true
    tty: true
    
    
