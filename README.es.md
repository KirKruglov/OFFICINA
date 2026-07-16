<p align="center">
  <a href="README.md">English</a> ·
  <a href="README.de.md">Deutsch</a> ·
  <strong>Español</strong> ·
  <a href="README.fr.md">Français</a> ·
  <a href="README.ru.md">Русский</a>
</p>

<h1 align="center">OFFICINA</h1>

<p align="center">
  <em>del latín</em> <strong>officina</strong> — un taller, una fragua donde se forjan herramientas.
</p>

<p align="center">
  <img alt="Plataforma: macOS" src="https://img.shields.io/badge/platform-macOS-000000?logo=apple&logoColor=white">
  <img alt="Núcleo: Claude Code" src="https://img.shields.io/badge/core-Claude%20Code-da7756">
  <img alt="Licencia: MIT" src="https://img.shields.io/badge/license-MIT-green">
  <a href="https://github.com/KirKruglov/OFFICINA/commits/main"><img alt="Último commit" src="https://img.shields.io/github/last-commit/KirKruglov/OFFICINA?logo=github&label=%C3%BAltimo%20commit"></a>
</p>

<p align="center">
  <strong>Método y herramientas para el trabajo sistemático con un asistente de desarrollo con IA.</strong>
  Skills, subagentes, CLI y metodología — reunidos y probados en proyectos reales.
</p>

---

> <sub>Traducido de la <a href="README.md">README en inglés</a> canónica; puede ir por detrás del original.</sub>

## Por qué OFFICINA

OFFICINA es un taller para trabajar con herramientas de IA en el desarrollo de productos. En su centro
está el **método**: cómo estructurar el trabajo con un asistente de IA para que sea predecible y
reproducible. A su alrededor — skills, subagentes, CLI y configuraciones, listos para reutilizar.

- **El método primero.** Reglas y guías reutilizables — estilo de escritura, creación de CLI, creación
  de skills y subagentes, ingeniería de bucles — cada una legible por sí sola.
- **Skills portables, Claude Code en el núcleo.** Los skills no están atados a una herramienta —
  funcionan en Claude Code y entornos compatibles. Los subagentes y ajustes están afinados para Claude
  Code. Ambos están seleccionados a mano en lugar de volcados en bloque.
- **El entorno alrededor del núcleo.** VS Code como sección aplicada que apoya el núcleo en lugar de
  ser el titular.

## Para quién

- **Trabajas con un asistente de IA y quieres sistema** — un método, no prompts sueltos.
- **Escribes tus propios skills, subagentes, CLI** — tomas patrones y guías probados.
- **Configuras tu entorno de trabajo** para desarrollo con IA en macOS.

Si buscas una distribución plug-and-play de «instalar y olvidar» — no es esto: OFFICINA va de método y
selección deliberada, no de un paquete llave en mano listo para usar.

## Qué incluye

| Sección | Qué contiene |
|---|---|
| [`skills/`](skills/) | Skills portables — modos y procedimientos reutilizables (Claude Code y entornos compatibles) |
| [`claude/`](claude/) | Capa de Claude Code — subagentes y ajustes |
| [`methodology/`](methodology/) | Reglas y guías — *cómo trabajar* |
| [`cli/`](cli/) | Herramientas CLI personales y bibliotecas compartidas |
| [`vscode/`](vscode/) | Ajustes de VS Code, atajos de teclado, extensiones curadas |
| [`install/`](install/) | Scripts de instalación y disposición |

## Cómo funciona el método

El punto de partida es una pregunta — con qué resolver la tarea:

- Una acción determinista, sin modelo — **CLI** (p. ej. [`jig`](cli/jig/) despliega un nuevo repositorio con un solo comando)
- Hay que entender el contexto y decidir según la situación — **skill**
- Una subtarea aislada o un rol-persona — **subagente**
- Un ciclo recurrente por horario — **loop**

Después viene la guía correspondiente en [`methodology/`](methodology/): estructura, convenciones, una
lista de verificación antes del despliegue. El método es común; el artefacto se traslada entre
proyectos.

[`jig`](cli/jig/) es el método convertido en herramienta: con un solo comando despliega la misma
estructura de documentos y configuración de Claude que describen las guías en [`methodology/`](methodology/).

## Inicio rápido

Requiere **macOS**. Clona el repositorio y abre la sección que se ajuste a tu necesidad — cada una trae
su propia README y, cuando corresponde, un instalador de un solo comando en [`install/`](install/):

```bash
git clone https://github.com/KirKruglov/OFFICINA.git
cd OFFICINA
```

El primer instalador listo configura el entorno de VS Code:

```bash
./install/vscode.sh   # ajustes, atajos, fuente MesloLGS Nerd Font, extensiones curadas
```

## Filosofía

- **La herramienta correcta para la tarea.** Lo determinista va a un script, las decisiones según
  contexto al modelo. Un modelo de más en el bucle añade coste e incertidumbre.
- **La fiabilidad se apoya en las restricciones.** Límites, presupuestos y comprobaciones sostienen el
  sistema. Un bucle autónomo necesita algo capaz de decir «no».
- **El conocimiento vive en el artefacto.** Una guía, skill o subagente se reutiliza; lo que se
  resuelve una vez no se resuelve de nuevo cada vez.
- **Seleccionado a mano.** Lo que llega al repositorio está probado en la práctica y elegido de forma
  deliberada.

## Contribuir

Se agradecen issues, preguntas y pequeñas correcciones — consulta [CONTRIBUTING.md](CONTRIBUTING.md).
OFFICINA es un proyecto abierto.

¿Te resulta útil? Dale una estrella — así el siguiente desarrollador encuentra el proyecto. Haz un
fork, toma lo que te sirva y adáptalo a tu configuración.

## Licencia

Publicado bajo la [Licencia MIT](LICENSE) — © 2026 Kir Kruglov. Libre de usar, modificar y distribuir.
