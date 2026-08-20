# Glossário

Um termo canônico por conceito, e os sinônimos que ninguém deve usar no lugar dele. Vale
em código, em requisito, no wiki e na conversa. O identificador em inglês entre parênteses
é como o conceito se chama no código — `.claude/rules/project.md` fixa essa convenção.

- **sala** (`room`) — o lugar onde a mídia acontece: até oito participantes e dezenas de espectadores, criada por quem hospeda a comunidade. Avoid: room, channel
- **participante** (`participant`) — quem está dentro da sala emitindo: câmera, microfone, e pode abrir a tela. Avoid: streamer
- **espectador** (`viewer`) — quem só recebe o que a sala emite, e pode ser promovido a participante. Avoid: viewer
- **anfitrião** (`host`) — o participante que comanda a sala: encerra, silencia, remove, decide quem entra. Avoid: host, moderador, broadcaster
- **comunidade** (`community`) — o conjunto de pessoas ligadas a um nó dono, e o limite de quem entra nas salas dele. Avoid: instância
- **nó** (`node`) — um processo do servidor rodando numa máquina da rede. Avoid: relay
- **nó dono** (`ownerNode`) — o nó que hospeda a comunidade, cria as salas dela e gera a chave de grupo
- **nó de passagem** (`relayNode`) — um nó que encaminha mídia de uma sala que não é dele, sem conseguir abri-la. Avoid: retransmissor, super-nó
- **plano de controle** (`controlPlane`) — o que decide: identidade, quem entra, em que sala, com que chave, em que nó
- **plano de dados** (`dataPlane`) — o que apenas encaminha bytes, e não decide nada
- **malha** (`mesh`) — mídia direta entre navegadores, sem servidor no caminho, enquanto a sala é pequena
- **rede** (`network`) — o conjunto de nós que encaminha a mídia quando a malha não cabe mais
- **árvore de distribuição** (`tree`) — o caminho que uma transmissão percorre de nó em nó, uma cópia por nó vizinho em vez de uma por espectador
- **promoção** (`promotion`) — a passagem de um estado ao seguinte sem cortar a chamada: da malha para a rede, ou de espectador a participante
- **camada** (`layer`) — uma das qualidades que um participante emite em paralelo, entre as quais o nó escolhe por assinante
- **bilhete** (`ticket`) — a autorização assinada que permite a um cliente usar um nó específico numa sala específica
- **token de anfitrião** (`hostToken`) — a autorização assinada que permite comandar uma sala, e nada além dela
- **chave de grupo** (`groupKey`) — a chave simétrica que o nó dono gera e entrega à comunidade, com que a mídia da sala é cifrada
- **geração de chave** (`keyGeneration`) — o número que avança a cada rotação da chave de grupo, e diz a que chave um quadro pertence
- **quadro cifrado** (`sealedFrame`) — o quadro de mídia como trafega: carga cifrada e o cabeçalho que o nó precisa para rotear
- **economia de banda** (`bandwidthSaving`) — quanto a árvore poupa na origem, medida contra a mesma sala servida por um nó só
