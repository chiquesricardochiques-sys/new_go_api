package professional_schedule

import (
	"database/sql"
	"fmt"
	"time"
)

type Service struct {
	repository *Repository
	db         *sql.DB
}

func NewService(
	repository *Repository,
	db *sql.DB,
) *Service {
	return &Service{
		repository: repository,
		db:         db,
	}
}

var diasValidos = map[string]bool{
	"seg": true,
	"ter": true,
	"qua": true,
	"qui": true,
	"sex": true,
	"sab": true,
	"dom": true,
}

func (s *Service) ObterGrade(
	salaoID uint64,
	profissionalID uint64,
) ([]ScheduleDay, error) {
	existe, err := s.repository.ProfissionalPertenceAoSalao(
		profissionalID,
		salaoID,
	)

	if err != nil {
		return nil, err
	}

	if !existe {
		return nil, fmt.Errorf(
			"profissional não pertence ao salão",
		)
	}

	return s.repository.ListarGrade(
		salaoID,
		profissionalID,
	)
}

// InicializarGrade cria a grade inicial do profissional.
//
// O profissional começa com os 7 dias cadastrados,
// porém todos fechados.
func (s *Service) InicializarGrade(
	salaoID uint64,
	profissionalID uint64,
) error {
	existe, err := s.repository.ProfissionalPertenceAoSalao(
		profissionalID,
		salaoID,
	)

	if err != nil {
		return err
	}

	if !existe {
		return fmt.Errorf(
			"profissional não pertence ao salão",
		)
	}

	tx, err := s.db.Begin()

	if err != nil {
		return err
	}

	if err := s.repository.InicializarGrade(
		tx,
		salaoID,
		profissionalID,
	); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Service) SalvarGrade(
	salaoID uint64,
	profissionalID uint64,
	grade []ScheduleDay,
) error {
	existe, err := s.repository.ProfissionalPertenceAoSalao(
		profissionalID,
		salaoID,
	)

	if err != nil {
		return err
	}

	if !existe {
		return fmt.Errorf(
			"profissional não pertence ao salão",
		)
	}

	if err := validarGrade(grade); err != nil {
		return err
	}

	tx, err := s.db.Begin()

	if err != nil {
		return err
	}

	if err := s.repository.SalvarGrade(
		tx,
		salaoID,
		profissionalID,
		grade,
	); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func validarGrade(
	grade []ScheduleDay,
) error {
	diasInformados := make(map[string]bool)

	for _, item := range grade {
		if !diasValidos[item.DiaSemana] {
			return fmt.Errorf(
				"dia da semana inválido: %s",
				item.DiaSemana,
			)
		}

		if diasInformados[item.DiaSemana] {
			return fmt.Errorf(
				"dia da semana duplicado: %s",
				item.DiaSemana,
			)
		}

		diasInformados[item.DiaSemana] = true

		// Dia fechado.
		//
		// Quando abre = false, não exigimos
		// nenhum horário.
		if !item.Abre {
			continue
		}

		if item.Abertura == "" {
			return fmt.Errorf(
				"abertura obrigatória para %s",
				item.DiaSemana,
			)
		}

		if item.Fechamento == "" {
			return fmt.Errorf(
				"fechamento obrigatório para %s",
				item.DiaSemana,
			)
		}

		abertura, err := parseHora(
			item.Abertura,
		)

		if err != nil {
			return fmt.Errorf(
				"abertura inválida para %s",
				item.DiaSemana,
			)
		}

		fechamento, err := parseHora(
			item.Fechamento,
		)

		if err != nil {
			return fmt.Errorf(
				"fechamento inválido para %s",
				item.DiaSemana,
			)
		}

		if !abertura.Before(fechamento) {
			return fmt.Errorf(
				"abertura deve ser menor que fechamento para %s",
				item.DiaSemana,
			)
		}

		// Sem almoço.
		if item.PausaInicio == "" &&
			item.PausaFim == "" {
			continue
		}

		// Se um dos dois foi informado,
		// os dois são obrigatórios.
		if item.PausaInicio == "" {
			return fmt.Errorf(
				"pausa_inicio obrigatória quando houver almoço para %s",
				item.DiaSemana,
			)
		}

		if item.PausaFim == "" {
			return fmt.Errorf(
				"pausa_fim obrigatória quando houver almoço para %s",
				item.DiaSemana,
			)
		}

		pausaInicio, err := parseHora(
			item.PausaInicio,
		)

		if err != nil {
			return fmt.Errorf(
				"pausa_inicio inválida para %s",
				item.DiaSemana,
			)
		}

		pausaFim, err := parseHora(
			item.PausaFim,
		)

		if err != nil {
			return fmt.Errorf(
				"pausa_fim inválida para %s",
				item.DiaSemana,
			)
		}

		if !abertura.Before(pausaInicio) {
			return fmt.Errorf(
				"abertura deve ser menor que pausa_inicio para %s",
				item.DiaSemana,
			)
		}

		if !pausaInicio.Before(pausaFim) {
			return fmt.Errorf(
				"pausa_inicio deve ser menor que pausa_fim para %s",
				item.DiaSemana,
			)
		}

		if !pausaFim.Before(fechamento) {
			return fmt.Errorf(
				"pausa_fim deve ser menor que fechamento para %s",
				item.DiaSemana,
			)
		}
	}

	return nil
}

func parseHora(
	value string,
) (time.Time, error) {
	formats := []string{
		"15:04",
		"15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(
			format,
			value,
		); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf(
		"hora inválida: %s",
		value,
	)
}

func (s *Service) ResetarGrade(
	salaoID uint64,
	profissionalID uint64,
) error {
	existe, err := s.repository.ProfissionalPertenceAoSalao(
		profissionalID,
		salaoID,
	)

	if err != nil {
		return err
	}

	if !existe {
		return fmt.Errorf(
			"profissional não pertence ao salão",
		)
	}

	tx, err := s.db.Begin()

	if err != nil {
		return err
	}

	if err := s.repository.ResetarGrade(
		tx,
		salaoID,
		profissionalID,
	); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}