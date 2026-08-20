---
status: accepted
---

# 0002 — Afinidade de sala por hash, e teto de nós por sala

## Context

Com mais de um nó, dois clientes da mesma sala podem chegar por nós diferentes, e cada
nó teria a sua versão de quem está dentro. Sala partida ao meio não é lentidão, é
defeito: os dois lados funcionam e não se veem.

A saída óbvia é um banco compartilhado com o estado das salas, o que acrescenta um
serviço com estado, sua replicação e sua operação — para guardar objetos que duram
minutos.

Há um segundo problema, de sinal oposto, que aparece quando a rede fica grande. Se cada
cliente for para o nó mais próximo dele, uma sala se espalha por tantos nós quantos
forem seus membros, e a árvore de distribuição vira malha completa entre nós. Nesse
ponto, acrescentar nó **piora** a rede: paga-se encaminhamento entre nós sem economizar
nada na origem.

## Decision

Cada sala tem um nó dono da sinalização, escolhido por **hash consistente do
identificador da sala** sobre o anel de nós. Qualquer nó que receba um cliente de uma
sala que não é sua o encaminha ao dono. O estado da sala vive na memória do dono, num
lugar só, e não há banco compartilhado.

Uma sala **não migra** quando o anel muda. Ela termina no nó que a hospeda; apenas salas
novas enxergam o anel novo.

Uma sala ocupa no máximo **três nós**. A proximidade geográfica é critério de desempate
entre nós elegíveis, nunca o critério principal: quando o teto é atingido, o participante
vai para um nó que a sala já ocupa, mesmo que não seja o mais próximo dele.

## Consequences

A queda do nó dono derruba a sinalização daquela sala, e ela precisa ser recriada. É a
troca deliberada por não ter estado compartilhado, e o custo é limitado porque salas são
objetos de vida curta.

Um participante pode ficar em um nó pior do que o mais próximo. A latência a mais de um
salto é menor do que a que a árvore espalhada acrescentaria, e o teto é o que mantém a
economia de banda real.

A árvore por sala é rasa por construção — no máximo três nós, portanto no máximo dois
saltos entre nós — o que mantém a latência vidro a vidro previsível.

Números como três nós por sala são de política, não de arquitetura, e mudam com medição
sem que este registro precise ser substituído.
