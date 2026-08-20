# Stack

Cada tecnologia adotada, com a linha que diz por que ela entrou. **Tecnologia que não
está aqui é decisão em aberto, nunca algo adotado em silêncio** — e como os manifestos
são dados estruturados, isso é verificável: dependência direta declarada lá e ausente
daqui é apontada.

Acrescentar dependência são dois atos: acrescentar, e dizer aqui por quê.

## Servidor

- **Go** — o nó precisa manter milhares de conexões simultâneas com latência previsível, e o modelo de goroutine mais o coletor de baixa pausa entregam isso sem o autor escrever laço de evento.
- **Pion** — pilha WebRTC em Go puro, sem `cgo` e sem `libwebrtc`, com acesso ao pacote RTP em vez de uma sessão opaca. É o que permite o nó encaminhar mídia que não consegue abrir.
- **biblioteca padrão de HTTP** — o roteamento por método e padrão que o Go traz desde a 1.22 cobre a API REST de sala. Framework é decisão a tomar quando doer, não antes.

## Cliente

- **TypeScript** — o protocolo tem estados com transição rígida — malha, rede, promoção — e tipo discriminado é o que impede um cliente de mandar mensagem que o servidor não espera.
- **Vite** — empacota a biblioteca e serve o protótipo com recarga rápida, sem configuração para um projeto desta forma.

## Desenvolvimento

- **Vitest** — roda os testes do cliente com a mesma configuração do Vite, sem um segundo pipeline de transformação para manter.
- **jsdom** — dá ao teste o `DOM` que o código de sala assume, sem subir navegador para verificar uma função.
- **ESLint** com **typescript-eslint** — a camada que encontra o que o compilador não encontra: promessa solta, variável não usada, retorno ignorado.
- **Prettier** — formatação decidida por ferramenta, para que revisão discuta o que a mudança faz.
- **golangci-lint** — o mesmo papel no lado Go, com `errcheck` e `nilerr`: erro não verificado é defeito que teste não encontra.
