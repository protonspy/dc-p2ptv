---
status: accepted
---

# 0001 — Planos de controle e de dados separados

## Context

A rede é federada: qualquer pessoa sobe um nó numa VPS e passa a encaminhar mídia de
salas que não são dela. Isso põe código de terceiro no caminho de dados de uma
comunidade que não o escolheu e não o audita.

Duas ameaças são diferentes e costumam ser tratadas como uma só. Um nó que **vê** a
mídia é um problema de confidencialidade, e cifra resolve. Um nó que **decide** — quem
entra, em que sala, com que papel, com que chave — é um problema de autoridade, e cifra
não resolve nada: um nó que admite participantes pode admitir a si mesmo.

A segunda é a pior das duas, e é a que passa despercebida, porque um nó que encaminha
naturalmente sabe muita coisa sobre a sala e é tentador deixá-lo responder por ela.

## Decision

O sistema tem dois planos, e a fronteira entre eles é a fronteira de confiança.

O **plano de controle** é central, operado por quem opera a rede, e é o único que
decide: identidade, associação à comunidade, criação e encerramento de sala, quem entra,
em que papel, em que nó, e a entrega da chave de grupo. Ele emite duas autorizações
assinadas — o bilhete, que permite a um cliente usar um nó numa sala, e o token de
anfitrião, que permite comandar uma sala e nada além dela.

O **plano de dados** é o conjunto de nós, voluntários e não confiáveis. Um nó verifica
a assinatura do bilhete que lhe é apresentado e encaminha. Ele não consulta o plano de
controle para decidir, porque não decide: a assinatura já carrega a decisão.

Um nó nunca admite participante, nunca cria sala, nunca conhece a chave de grupo de uma
sala que não é da comunidade dele, e nunca é consultado sobre autorização.

## Consequences

O bilhete precisa ser verificável offline e ter validade curta, porque revogar algo que
não se consulta é impossível: a janela de validade *é* a revogação.

A entrada numa sala passa a custar uma ida ao plano de controle antes da mídia começar.
Isso acrescenta latência ao ingresso — não ao fluxo — e é o preço de o nó não decidir.

O plano de controle é ponto único de falha para **entrar** numa sala. Salas em andamento
sobrevivem à queda dele, porque os bilhetes já emitidos continuam válidos até expirar;
ninguém novo entra enquanto ele estiver fora.

Um nó que passe a decidir alguma coisa é um defeito de arquitetura, não uma otimização.
Revisão de código tem uma pergunta fixa por causa disto: este caminho está decidindo, ou
apenas encaminhando?
