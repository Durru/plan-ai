# REGLA INVIOLABLE

Este VPS es **exclusivamente para CONSTRUIR** plan-ai. NO para probarlo.

## Prohibido

- ❌ **NUNCA** ejecutar `plan-ai install`, `plan-ai init`, `plan-ai sync`, `plan-ai update`, `plan-ai uninstall`, `plan-ai setup opencode`, `plan-ai bootstrap`, ni ningún comando de plan-ai que escriba estado o configuración en este VPS.
- ❌ **NUNCA** escribir en `~/.config/opencode/` ni `~/.opencode/` desde código de plan-ai.
- ❌ **NUNCA** usar plan-ai dentro de su propio proyecto (`/root/plan-ai/`).
- ❌ **NUNCA** modificar `opencode.json`, `config.json`, ni ningún archivo de configuración de OpenCode desde plan-ai en este VPS.

## Permitido

- ✅ `go build ./...` — compilar
- ✅ `go test ./...` — tests unitarios (usan sandbox aislado automáticamente)
- ✅ `go vet ./...` — análisis estático
- ✅ `go test -count=1 ./...` — tests sin caché

## Cómo hacer pruebas sin romper nada

- **Tests unitarios**: los tests de Go usan `t.TempDir()` y `OPENCODE_CONFIG_DIR` automáticamente en sandbox. No tocan el sistema real.
- **Tests de integración**: configurar `OPENCODE_CONFIG_DIR=/tmp/sandbox-oc` y `HOME=/tmp/sandbox-home` antes de cualquier comando de plan-ai.
- **Pruebas reales**: usar el VPS de pruebas dedicado. **Nunca este VPS**.

## Por qué existe esta regla

Una sola ejecución de `plan-ai install` o `SetupMCPConfig` sobrescribió `~/.opencode/config.json` y `~/.config/opencode/config.json`, borrando la sección `provider` y dejando OpenCode inservible por horas. No vuelve a pasar.
