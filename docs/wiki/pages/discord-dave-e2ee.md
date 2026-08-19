# DAVE — E2EE de áudio e vídeo do Discord

DAVE (*Discord Audio & Video End-to-End Encryption*) é a camada que tira o servidor da
lista de quem consegue ler a mídia. Combina **MLS** (RFC 9420) para acordo de chave em
grupo com a **WebRTC Encoded Transform API** para cifrar quadro a quadro depois de
codificar e antes de empacotar. Ativo desde setembro de 2024 em DMs, chamadas de grupo,
canais de voz de servidor (exceto palco) e Go Live; auditado pela Trail of Bits.

Ortogonal à cifra de transporte de [[discord-transporte-de-midia]]: aquela protege o
fio, esta protege contra o próprio SFU.

## MLS

Ciphersuite `DHKEMP256_AES128GCM_SHA256_P256`. O voice gateway atua como **external
sender** e como serviço de entrega e autenticação — só ele pode propor *add* e *remove*,
e nenhuma outra proposta é aceita.

O fluxo de entrada: o membro pendente envia seu **key package**; o gateway distribui a
proposta de adição aos membros atuais; membros mandam **commits** referenciando as
propostas; o gateway **escolhe o primeiro commit válido de cada epoch** e o difunde.
Escolher um só commit é o que impede conluio entre membros que commitam. O novo membro
recebe um **Welcome** direcionado com os segredos do grupo.

Na remoção, o removido perde a capacidade de decifrar assim que o novo epoch executa —
com uma janela de cerca de 2 segundos em que ainda decifra mídia em voo. Se sobra um
único membro, o grupo é reiniciado, para que ninguém fique sozinho no controle do
estado.

## Chaves e rotação

Cada epoch deriva chaves simétricas **por emissor** a partir dos exporters do MLS.
Dentro de um epoch, o emissor faz *ratchet* quando o contador de nonce chega a 2^24
quadros — o byte mais significativo do nonce é a geração. Chaves do epoch anterior ficam
em cache por cerca de 10 segundos, para decifrar o que estava em trânsito.

## Formato do quadro

AES-128-GCM com tag truncada em 64 bits. O quadro sai com trechos em claro intercalados
com trechos cifrados — o que o codec exige ler antes de decodificar fica exposto — e,
no fim, um rodapé:

```
[ ...dados... ][ tag 8B ][ nonce ULEB128 ][ pares (offset,tamanho) das faixas em
claro, ULEB128 ][ tamanho do suplemento 1B ][ marcador mágico 0xFAFA ]
```

O **cifrador é ciente do codec** (Opus, VP8, VP9, H.264, H.265, AV1 têm exigências
diferentes de cabeçalho em claro); o **decifrador é agnóstico** — ele lê o rodapé e não
precisa saber o codec. Boa decisão de desenho: um codec novo muda só um lado.

## Passthrough e transições

Quadro que falha na checagem de quadro-de-protocolo (sem marcador, ou suplemento
malformado) passa direto, sem cifrar nem decifrar. É isso que faz sessões mistas e
transições funcionarem sem cortar o áudio.

O gateway anuncia mudanças com `dave_protocol_prepare_transition`; os clientes entram em
passthrough e respondem *transition ready*; então vem `dave_protocol_execute_transition`.
Mudança de versão do protocolo pode exigir recriar o grupo MLS no epoch 1 ou apenas
avançar o epoch. Quando entra um cliente que não suporta DAVE, a sessão **rebaixa** para
cifra só de transporte, e volta a subir quando todos suportam. Durante o rebaixamento o
SFU filtra quadros de protocolo dos clientes que não suportam, para não injetar lixo no
decodificador.

O whitepaper reconhece um problema aberto: pacotes de silêncio (`0xF8FFFE`) passam mesmo
sob E2EE, para que o SFU consiga sintetizar mudo — e isso é um vetor de negação de
serviço em correção.

## Verificação

Fingerprint par a par derivado por scrypt sobre as chaves públicas, os IDs de usuário e
a versão, comparado fora de banda. Chaves de assinatura persistentes sobrevivem à
sessão onde a plataforma permite guardá-las com segurança; caso contrário, efêmeras por
sessão. O **epoch authenticator** do MLS é exibido como código de 30 dígitos para
verificação do grupo inteiro.

## Consequência para redistribuição

Sob DAVE, quem não é membro do grupo MLS não decifra nada — nem o SFU. Um redistribuidor
P2P que não seja um cliente legítimo dentro do grupo recebe bytes opacos. Ver
[[viabilidade-p2ptv-sobre-discord]].
