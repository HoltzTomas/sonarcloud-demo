# Contract API — demo de remediación con SonarQube/SonarCloud

Repo de demo para el flujo "el scanner encuentra, Devin triage y remedia".
Simula el sistema de gestión de contratos geoposicionados: **app web
en Go** (UI + datos seed en SQLite en memoria) + infraestructura Terraform, con
vulnerabilidades plantadas y **tres falsos positivos deliberados**, que es donde
está el diferencial frente al scanner.

Lo importante: casi todos los hallazgos son **explotables desde el navegador**,
así que cada sesión de Devin puede mostrar el exploit y el fix en la UI en vez
de argumentarlos por escrito.

## Levantar la app

```
go run ./cmd/api      # http://localhost:8080
```

Sin base de datos externa ni configuración: los contratos se siembran en memoria
al arrancar.

| Página | Ruta |
| --- | --- |
| Buscar contratos | `/` |
| Detalle de contrato | `/contracts/C-1001` |
| Descargar adjunto | `/contracts/attachment?name=C-1001.txt` |
| Exportar a PDF | `/contracts/export?id=C-1001` |
| Resumen por región | `/summary?region=ES-01` |

## Qué hay plantado

| Archivo | Hallazgo | Explotable desde | Real / FP |
| --- | --- | --- | --- |
| `internal/contracts/repository.go` (`FindByCustomer`) | SQL injection (`fmt.Sprintf` en el query) | buscador: `' OR '1'='1` devuelve todos los contratos | real |
| `internal/web/handlers.go` (`Search`) | XSS reflejado (`template.HTML` sobre el input) | buscador: `<script>alert(1)</script>` ejecuta | real |
| `internal/web/handlers.go` (`Attachment`) | Path traversal | `?name=../private/tenant-keys.txt` devuelve claves internas | real |
| `internal/web/handlers.go` (`Export`) | Command injection (`sh -c`) | `?id=C-1001; echo PWNED_$(id -u)` ejecuta comandos | real |
| `internal/auth/tokens.go` | Secretos hardcodeados + MD5 para passwords | código | real |
| `internal/integrations/gis_client.go` | `InsecureSkipVerify: true` (TLS sin validar) | código | real |
| `infra/main.tf` | Bucket público, SG 0.0.0.0/0, RDS sin cifrar y público, password en claro | IaC | real |
| `internal/contracts/repository.go` (`CountByValidatedRegion`) | Sonar marca SQL injection por la concatenación, pero `/summary` valida la región contra `^[A-Z]{2}-[0-9]{2}$` antes de llamar | el payload devuelve `400 invalid region code` | **falso positivo** |
| `internal/cache/key.go` | MD5 marcado como criptografía débil, pero es una clave de cache local | — | **falso positivo** |
| `internal/fixtures/fixtures.go` | "Credenciales" que sólo existen en el stack de tests efímero | — | **falso positivo** |

`CountByRegion` en `repository.go` ya usa query parametrizada: sirve para
mostrar el contraste con la línea vulnerable de al lado.

El falso positivo de `/summary` es el mejor momento de la demo: **parece** la
misma SQL injection que la del buscador — misma concatenación, misma regla de
Sonar — y en el navegador el mismo payload que en `/` filtra toda la base, acá
rebota con un 400.

## Flujo de la demo

1. CI corre el scan de SonarCloud sobre el PR y el quality gate falla
   (`.github/workflows/sonarcloud.yml`).
2. Al fallar, `.github/workflows/devin-remediate.yml` crea **una sesión
   orquestadora** con `playbooks/sonar-gate-owner.md`, que recibe los findings
   y las condiciones de gate en error.
3. La sesión orquestadora crea sub-sesiones por finding con el playbook de
   triage, clasifica cada caso como real o falso positivo y se ocupa también de
   las condiciones sin finding (como coverage), siempre en la misma rama del
   PR y sin abrir PRs nuevos.
4. Cada push re-dispara el scan; la orquestadora itera hasta que el quality gate
   queda verde.

Alternativa sin API: la automation template **SonarQube Quality Gate Fix**
(Automations en la webapp) hace el paso 2 sin escribir workflow propio. El
script existe porque en cuentas como el disparo sale del pipeline y hay
que mostrar la orquestación de findings y condiciones del gate.

## Setup

1. Crear el proyecto en SonarCloud dentro de la org `HoltzTomas`
   (`https://sonarcloud.io/organizations/HoltzTomas/projects`), con project key
   `HoltzTomas_sonarcloud-demo`. Desactivar "Automatic Analysis" para que
   corra el scanner de CI.
2. En el repo de GitHub, cargar los secrets `SONAR_TOKEN` (token de SonarCloud)
   y `DEVIN_API_KEY` (API key de la org de Devin dueña del repo).
3. Abrir un PR con cualquier cambio menor sobre `main` para disparar el scan y
   ver el gate en rojo.

## Talk track (~10 min)

- **0-1 min** — El PR con el quality gate en rojo. "Esto es lo que ya tienen hoy:
  el scanner les dice qué está mal, y ahí se termina."
- **1-3 min** — La lista de findings en Sonar, mezclando reales y ruido. "Acá es
  donde hoy se van 5-6 personas revisando falsos positivos durante semanas."
- **3-5 min** — La sesión orquestadora y sus sub-sesiones de triage por
  hallazgo, incluyendo una condición de coverage sin finding.
- **5-8 min** — Abrir dos sesiones y mostrar sus grabaciones: una que **explota
  y arregla** la SQL injection del buscador (payload en la pantalla, todos los
  contratos filtrados, fix, mismo payload sin efecto), y otra que **descarta** la
  SQL injection de `/summary` — misma regla, mismo tipo de código — mostrando el
  ataque rebotando contra el validador. Ese contraste es el argumento entero de
  la demo.
- **8-10 min** — El mismo PR con los fixes commiteados por Devin, los
  comentarios de triage y el gate en verde. Cierre:
  "no reemplazamos a Sonar, le ponemos la capa de inteligencia y ejecución
  encima".

## Adaptación

- El input real son Fortify y OX además de Sonar: el script sólo cambia en la
  función `findings()`; el playbook de triage es el mismo.
- El destino de los PRs es Azure DevOps Repos, no GitHub — mostrar el flujo con
  ADO si ya está el acceso, o aclararlo verbalmente si la demo va sobre GitHub.
- El backtest que pidieron (comparar el triage de Devin contra el manual en 2-3
  apps) es exactamente el paso 3 corrido en batch: conviene mencionarlo cuando
  se muestre la sesión del falso positivo.
- La verificación en navegador es el gancho para sus apps con frontend: el mismo
  playbook sirve para .NET o Salesforce cambiando sólo el comando de arranque.
