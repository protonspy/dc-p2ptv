# Captura e transmissão de tela

Tela não é câmera. O conteúdo é majoritariamente estático, com bordas duras, texto fino
e áreas de cor chapada, e depois muda tudo de uma vez quando alguém rola a página.
Codificador ajustado para rosto falando produz texto borrado; ajustado para texto, gasta
banda demais em vídeo dentro da tela.

## Captura no navegador

```ts
const stream = await navigator.mediaDevices.getDisplayMedia({
  video: { displaySurface: "monitor", frameRate: { ideal: 30, max: 60 } },
  audio: true,
});
const [track] = stream.getVideoTracks();
track.contentHint = "text"; // "detail" | "text" | "motion"
```

- **`contentHint`** é o botão mais importante e o mais esquecido. `text` e `detail`
  dizem ao codificador que nitidez vale mais que fluidez — ele preserva bordas e
  sacrifica taxa de quadros sob pressão. `motion` é o oposto, e é o padrão da câmera.
  Sem definir, compartilhamento de slides e de código sai ilegível.
- **`displaySurface`** sugere o tipo de superfície: `monitor`, `window`, `browser`,
  `application`. É sugestão — quem decide é o usuário no diálogo do sistema.
- **`cursor`**: `always`, `motion`, `never`.
- **`degradationPreference`** no lado do emissor: `maintain-resolution` para tela,
  `maintain-framerate` para câmera. Mesma lógica do `contentHint`, aplicada ao
  transporte.

## Áudio do sistema é o problema difícil

`getDisplayMedia({ audio: true })` **não** entrega áudio do sistema de forma uniforme.
No Chrome funciona para aba e, em parte, para tela no Windows; em muitos ambientes lança
`NotSupportedError`. No Firefox e no Safari, praticamente não existe.

Em aplicativo Electron o caminho atual é interceptar no processo principal:

```js
session.defaultSession.setDisplayMediaRequestHandler((request, callback) => {
  callback({ video: fonteEscolhida, audio: "loopback" });
});
```

O `desktopCapturer` antigo tem histórico ruim de loopback e de captura silenciosa, com
regressões documentadas de fluxo ativo sem som depois de atualizações. Se áudio do
sistema é requisito, trate como risco de plataforma, não como detalhe — é uma das razões
pelas quais o Discord mantém cliente nativo com media engine próprio
([[discord-arquitetura-de-midia]]) e uma flag separada de áudio de contexto, o
`SOUNDSHARE` de [[discord-voice-gateway]].

## Codec

**AV1** tem ferramentas de *screen content coding* — IntraBC e modo paleta — feitas
exatamente para este conteúdo, e a redução de bitrate com qualidade melhor é grande
contra VP9. Custo: codificar em tempo real é caro e a aceleração em hardware é desigual;
o Safari não expõe codificação AV1 pelo WebRTC.

Na prática, escada de codec: **AV1 onde o dispositivo aguenta, VP9 como alternativa,
H.264 como piso universal**. É a mesma ordem de preferência que o Discord anuncia em
[[discord-transporte-de-midia]].

Resolução importa mais que taxa de quadros aqui: 1080p a 15 quadros por segundo é
legível, 720p a 60 não é, quando o conteúdo é código.

## Captura em Go

Não há equivalente do `getDisplayMedia`; é integração por sistema operacional.

- **`kbinani/screenshot`** — X11 e Windows sem cgo, macOS com cgo. É a peça que a
  própria discussão do `pion/mediadevices` aponta como caminho para entrada de tela.
- **Windows** — a API que interessa é a **Desktop Duplication (DXGI)**: captura na
  composição, sem custo de leitura de janela. As implementações de referência são em C++
  e Python, como DXcam e D3DShot; em Go é binding.
- **Linux** — X11 direto, ou **PipeWire** com o portal `xdg-desktop-portal`, que é o
  único caminho no Wayland.
- **GStreamer**, com `d3d11screencapturesrc` e afins, resolve captura e codificação de
  uma vez, ao custo de trazer GStreamer para dentro do projeto.

Depois de capturar, o quadro precisa virar RTP. O `pion/mediadevices` cobre o caminho,
mas a codificação sai de cgo — ver a ressalva final de [[p2p-em-go]].

## Transmitir tela para muitos

Vale a distinção de [[player-de-streaming]]: tela **interativa**, com alguém
apresentando e respondendo perguntas, é WebRTC, sub-segundo, com SFU. Tela **assistida**
por muita gente, sem interação, é segmento e enxame — e aí os 2 segundos do MSE valem a
descarga de [[cdn-p2p-hibrido]].
