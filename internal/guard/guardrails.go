package guard

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/ccastromar/aos-banking-v2/internal/config"
)

// ---- helpers internos ----

func isValidPhoneNumber(s string) bool {
	re := regexp.MustCompile(`^[0-9+][0-9]{5,14}$`)
	return re.MatchString(s)
}

// Intent puede usar herramientas peligrosas?
func ValidateIntentPermissions(intent config.Intent, pipeline config.Pipeline, tools map[string]config.Tool) error {
	for _, step := range pipeline.Steps {

		// ⛔ Si el step no tiene tool → lo saltamos
		if step.Tool == "" {
			continue
		}

		t, ok := tools[step.Tool]
		if !ok {
			return fmt.Errorf("tool %s no encontrada", step.Tool)
		}

		if t.Mode == "dangerous" && !intent.AllowDangerous {
			return fmt.Errorf("el intent '%s' no permite tools peligrosas (tool=%s)", intent.Type, t.Name)
		}
	}
	return nil
}

// Validar params sensibles (amount, phone…)
func ValidateDangerousParams(intent config.Intent, params map[string]string) error {
	if !intent.AllowDangerous {
		return nil
	}

	if intent.RequiresAmount {
		raw := params["amount"]
		if raw == "" {
			return fmt.Errorf("falta parámetro requerido: amount")
		}
		amount, err := strconv.ParseFloat(raw, 64)
		if err != nil || amount <= 0 {
			return fmt.Errorf("amount inválido: %s", raw)
		}
		if intent.MaxAmount > 0 && amount > intent.MaxAmount {
			return fmt.Errorf("amount excede límite permitido: %v > %v", amount, intent.MaxAmount)
		}
	}

	if intent.RequiresPhone {
		phone := params["toPhone"]
		if phone == "" {
			return fmt.Errorf("falta parámetro requerido: toPhone")
		}
		if !isValidPhoneNumber(phone) {
			return fmt.Errorf("toPhone no válido: %s", phone)
		}
	}

	return nil
}

// No permitir dangerous → dangerous encadenado
func ValidateDangerousChain(pipeline config.Pipeline, tools map[string]config.Tool) error {
	dangerousSeen := false

	for _, step := range pipeline.Steps {

		// ⛔ Steps sin tool -> ignóralos
		if step.Tool == "" {
			continue
		}

		t, ok := tools[step.Tool]
		if !ok {
			return fmt.Errorf("tool %s no encontrada en chain check", step.Tool)
		}

		if t.Mode == "dangerous" {
			if dangerousSeen {
				return fmt.Errorf("pipeline '%s' encadena tools peligrosas", pipeline.Name)
			}
			dangerousSeen = true
		}
	}
	return nil
}

// ---- API pública: un solo punto de entrada ----

func ValidateAll(intent config.Intent, pipeline config.Pipeline, params map[string]string, tools map[string]config.Tool) error {
	if err := ValidateIntentPermissions(intent, pipeline, tools); err != nil {
		return err
	}
	if err := ValidateDangerousParams(intent, params); err != nil {
		return err
	}
	if err := ValidateDangerousChain(pipeline, tools); err != nil {
		return err
	}
	return nil
}
