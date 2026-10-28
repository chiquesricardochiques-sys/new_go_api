package profissional

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/url"
	"strings"
	"time"

	// Embute o banco de fusos horários no binário (necessário no Windows,
	// onde time.LoadLocation("America/Sao_Paulo") falha sem isso).
	_ "time/tzdata"
)

const (
	layoutData           = "2006-01-02"
	layoutSQL            = "2006-01-02 15:04:05"
	intervaloSlotMinutos = 30
)

var statusValidos = map[string]bool{
	"pendente":       true,
	"confirmado":     true,
	"concluido":      true,
	"cancelado":      true,
	"nao_compareceu": true,
}

var fusoBrasil = func() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("BRT", -3*60*60)
	}
	return loc
}()

// ClienteLookup busca nome/telefone dos clientes (que moram no Auth).
// Opcional: com nil, a agenda devolve só cliente_id.
type ClienteLookup interface {
	Buscar(ctx context.Context, authUserIDs []int64) (map[int64]ClienteInfo, error)
}

type Service struct {
	repository *Repository
	clientes   ClienteLookup
}

func NewService(repository *Repository, clientes ClienteLookup) *Service {
	return &Service{repository: repository, clientes: clientes}
}

// ============================================================================
// HELPERS
// ============================================================================

func agora() time.Time {
	return time.Now().In(fusoBrasil)
}

func inicioDoDia(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, fusoBrasil)
}

func parseData(valor string) (time.Time, error) {
	t, err := time.ParseInLocation(layoutData, strings.TrimSpace(valor), fusoBrasil)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: data deve estar no formato YYYY-MM-DD", ErrDadosInvalidos)
	}
	return t, nil
}

func diaSemanaSigla(t time.Time) string {
	return [...]string{"dom", "seg", "ter", "qua", "qui", "sex", "sab"}[int(t.Weekday())]
}

// inicioDaSemana devolve a segunda-feira da semana de t.
func inicioDaSemana(t time.Time) time.Time {
	dia := inicioDoDia(t)
	return dia.AddDate(0, 0, -((int(dia.Weekday()) + 6) % 7))
}

func minutos(hhmm string) int {
	var h, m int
	fmt.Sscanf(hhmm, "%d:%d", &h, &m)
	return h*60 + m
}

func horaString(totalMinutos int) string {
	return fmt.Sprintf("%02d:%02d", totalMinutos/60, totalMinutos%60)
}

func arredondar(v float64) float64 {
	return math.Round(v*100) / 100
}

func normalizar(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func validarBase(b BaseRequest) error {
	if b.IDSalao <= 0 {
		return fmt.Errorf("%w: id_salao obrigatório", ErrDadosInvalidos)
	}
	if b.ProfissionalID <= 0 {
		return fmt.Errorf("%w: profissional_id obrigatório", ErrDadosInvalidos)
	}
	return nil
}

// profissionalDoSalao valida os IDs e garante que o profissional pertence ao salão.
func (s *Service) profissionalDoSalao(ctx context.Context, b BaseRequest) (*Profissional, error) {
	if err := validarBase(b); err != nil {
		return nil, err
	}
	return s.repository.BuscarPorID(ctx, b.IDSalao, b.ProfissionalID)
}

func linkWhatsApp(telefone, nome string) *string {
	var b strings.Builder
	for _, r := range telefone {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}

	numeros := b.String()
	if numeros == "" {
		return nil
	}
	if len(numeros) == 10 || len(numeros) == 11 {
		numeros = "55" + numeros
	}

	msg := strings.ReplaceAll(url.QueryEscape("Olá "+nome+", tudo bem?"), "+", "%20")
	link := "https://wa.me/" + numeros + "?text=" + msg

	return &link
}

func (s *Service) enriquecerClientes(ctx context.Context, itens []AgendamentoItem) {
	if s.clientes == nil || len(itens) == 0 {
		return
	}

	vistos := make(map[int64]bool)
	ids := make([]int64, 0, len(itens))
	for _, it := range itens {
		if !vistos[it.ClienteID] {
			vistos[it.ClienteID] = true
			ids = append(ids, it.ClienteID)
		}
	}

	infos, err := s.clientes.Buscar(ctx, ids)
	if err != nil {
		// A agenda continua funcionando, só sem nome/telefone.
		log.Printf("[profissional] falha ao buscar clientes no Auth: %v", err)
		return
	}

	for i := range itens {
		info, ok := infos[itens[i].ClienteID]
		if !ok {
			continue
		}
		if info.Nome != "" {
			itens[i].Cliente = info.Nome
		}
		if info.Telefone != "" {
			tel := info.Telefone
			itens[i].ClienteTelefone = &tel
			itens[i].WhatsappLink = linkWhatsApp(tel, itens[i].Cliente)
		}
	}
}

// ============================================================================
// ME / PERFIL
// ============================================================================

// Me substitui o login-profissional: o app autentica no Auth e depois chama aqui
// com o auth_user_id para descobrir qual profissional é (e se está liberado).
func (s *Service) Me(ctx context.Context, req MeRequest) (PerfilResponse, error) {
	if req.IDSalao <= 0 {
		return PerfilResponse{}, fmt.Errorf("%w: id_salao obrigatório", ErrDadosInvalidos)
	}
	if req.AuthUserID <= 0 {
		return PerfilResponse{}, fmt.Errorf("%w: auth_user_id obrigatório", ErrDadosInvalidos)
	}

	p, err := s.repository.BuscarPorAuthUser(ctx, req.IDSalao, req.AuthUserID)
	if err != nil {
		return PerfilResponse{}, err
	}

	if p.Status != "ativo" {
		return PerfilResponse{}, ErrProfissionalInativo
	}

	return PerfilResponse{Profissional: *p}, nil
}

func (s *Service) Perfil(ctx context.Context, req BaseRequest) (PerfilResponse, error) {
	p, err := s.profissionalDoSalao(ctx, req)
	if err != nil {
		return PerfilResponse{}, err
	}
	return PerfilResponse{Profissional: *p}, nil
}

// ============================================================================
// AGENDA
// ============================================================================

func (s *Service) Agenda(ctx context.Context, req AgendaRequest) (AgendaResponse, error) {
	prof, err := s.profissionalDoSalao(ctx, req.BaseRequest)
	if err != nil {
		return AgendaResponse{}, err
	}

	dia := inicioDoDia(agora())
	if strings.TrimSpace(req.Data) != "" {
		dia, err = parseData(req.Data)
		if err != nil {
			return AgendaResponse{}, err
		}
	}

	status := normalizar(req.Status)
	if status == "todos" {
		status = ""
	}
	if status != "" && !statusValidos[status] {
		return AgendaResponse{}, fmt.Errorf("%w: status inválido", ErrDadosInvalidos)
	}

	itens, err := s.repository.ListarAgendamentos(ctx, FiltroAgendamentos{
		IDSalao:        req.IDSalao,
		ProfissionalID: prof.ID,
		Inicio:         dia.Format(layoutSQL),
		Fim:            dia.AddDate(0, 0, 1).Format(layoutSQL),
		Status:         status,
	})
	if err != nil {
		return AgendaResponse{}, err
	}

	for i := range itens {
		itens[i].UsaHorarios = prof.UsaHorarios
	}
	s.enriquecerClientes(ctx, itens)

	return AgendaResponse{
		Agendamentos: itens,
		Data:         dia.Format(layoutData),
	}, nil
}

// ============================================================================
// ALTERAR STATUS
// ============================================================================

func (s *Service) AlterarStatus(ctx context.Context, req AlterarStatusRequest) (AlterarStatusResponse, error) {
	if err := validarBase(req.BaseRequest); err != nil {
		return AlterarStatusResponse{}, err
	}
	if req.AgendamentoID <= 0 {
		return AlterarStatusResponse{}, fmt.Errorf("%w: agendamento_id obrigatório", ErrDadosInvalidos)
	}

	novoStatus := normalizar(req.NovoStatus)
	if !statusValidos[novoStatus] {
		return AlterarStatusResponse{}, fmt.Errorf("%w: status inválido", ErrDadosInvalidos)
	}

	// Garante que o agendamento é desse profissional/salão (o UPDATE do MySQL
	// devolve 0 linhas quando o status não muda, então não dá pra confiar nele).
	if _, err := s.repository.StatusDoAgendamento(ctx, req.IDSalao, req.ProfissionalID, req.AgendamentoID); err != nil {
		return AlterarStatusResponse{}, err
	}

	if err := s.repository.AtualizarStatus(ctx, req.IDSalao, req.ProfissionalID, req.AgendamentoID, novoStatus); err != nil {
		return AlterarStatusResponse{}, err
	}

	return AlterarStatusResponse{NovoStatus: novoStatus}, nil
}

// ============================================================================
// RESUMO
// ============================================================================

func (s *Service) Resumo(ctx context.Context, req BaseRequest) (ResumoResponse, error) {
	prof, err := s.profissionalDoSalao(ctx, req)
	if err != nil {
		return ResumoResponse{}, err
	}

	hoje := inicioDoDia(agora())
	semanaIni := inicioDaSemana(hoje)
	mesIni := time.Date(hoje.Year(), hoje.Month(), 1, 0, 0, 0, 0, fusoBrasil)

	contar := func(ini, fim time.Time) (int, error) {
		return s.repository.ContarPeriodo(ctx, req.IDSalao, prof.ID, ini.Format(layoutSQL), fim.Format(layoutSQL))
	}

	var resp ResumoResponse

	if resp.TotalHoje, err = contar(hoje, hoje.AddDate(0, 0, 1)); err != nil {
		return ResumoResponse{}, err
	}
	if resp.TotalSemana, err = contar(semanaIni, semanaIni.AddDate(0, 0, 7)); err != nil {
		return ResumoResponse{}, err
	}
	if resp.TotalMes, err = contar(mesIni, mesIni.AddDate(0, 1, 0)); err != nil {
		return ResumoResponse{}, err
	}

	return resp, nil
}

// ============================================================================
// HISTÓRICO
// ============================================================================

func (s *Service) Historico(ctx context.Context, req HistoricoRequest) (HistoricoResponse, error) {
	prof, err := s.profissionalDoSalao(ctx, req.BaseRequest)
	if err != nil {
		return HistoricoResponse{}, err
	}

	hoje := inicioDoDia(agora())
	var ini, fim string

	switch normalizar(req.Periodo) {
	case "":
		// sem filtro: tudo
	case "dia":
		ini, fim = hoje.Format(layoutSQL), hoje.AddDate(0, 0, 1).Format(layoutSQL)
	case "semana":
		sem := inicioDaSemana(hoje)
		ini, fim = sem.Format(layoutSQL), sem.AddDate(0, 0, 7).Format(layoutSQL)
	case "mes":
		mes := time.Date(hoje.Year(), hoje.Month(), 1, 0, 0, 0, 0, fusoBrasil)
		ini, fim = mes.Format(layoutSQL), mes.AddDate(0, 1, 0).Format(layoutSQL)
	default:
		return HistoricoResponse{}, fmt.Errorf("%w: periodo deve ser dia, semana ou mes", ErrDadosInvalidos)
	}

	itens, err := s.repository.ListarAgendamentos(ctx, FiltroAgendamentos{
		IDSalao:        req.IDSalao,
		ProfissionalID: prof.ID,
		Inicio:         ini,
		Fim:            fim,
		Desc:           true,
	})
	if err != nil {
		return HistoricoResponse{}, err
	}

	for i := range itens {
		itens[i].UsaHorarios = prof.UsaHorarios
	}
	s.enriquecerClientes(ctx, itens)

	// Como no Node, os totais consideram todo o histórico (não só o período).
	concluidos, cancelados, err := s.repository.TotaisHistorico(ctx, req.IDSalao, prof.ID)
	if err != nil {
		return HistoricoResponse{}, err
	}

	return HistoricoResponse{
		TotalConcluidos: concluidos,
		TotalCancelados: cancelados,
		Historico:       itens,
	}, nil
}

// ============================================================================
// HORÁRIOS DO DIA
// ============================================================================

type intervaloMin struct{ ini, fim int }

func bloqueiosEmMinutos(bloqueios []BloqueioHorario, dia time.Time) []intervaloMin {
	base := time.Date(dia.Year(), dia.Month(), dia.Day(), 0, 0, 0, 0, time.UTC)

	limitar := func(v int) int {
		if v < 0 {
			return 0
		}
		if v > 24*60 {
			return 24 * 60
		}
		return v
	}

	resultado := make([]intervaloMin, 0, len(bloqueios))
	for _, b := range bloqueios {
		ini, err1 := time.Parse("2006-01-02 15:04", b.Inicio)
		fim, err2 := time.Parse("2006-01-02 15:04", b.Fim)
		if err1 != nil || err2 != nil {
			continue
		}
		resultado = append(resultado, intervaloMin{
			ini: limitar(int(ini.Sub(base).Minutes())),
			fim: limitar(int(fim.Sub(base).Minutes())),
		})
	}

	return resultado
}

func gerarSlots(g *Grade, ocupados []AgendamentoItem, bloqueios []intervaloMin) []Slot {
	slots := make([]Slot, 0)

	abertura := minutos(g.Abertura)
	fechamento := minutos(g.Fechamento)

	pausaIni, pausaFim := -1, -1
	if g.PausaInicio != "" && g.PausaFim != "" {
		pausaIni, pausaFim = minutos(g.PausaInicio), minutos(g.PausaFim)
	}

	for m := abertura; m < fechamento; m += intervaloSlotMinutos {
		if pausaIni >= 0 && m >= pausaIni && m < pausaFim {
			continue
		}
		slots = append(slots, Slot{Hora: horaString(m), Disponivel: true})
	}

	for i := range slots {
		m := minutos(slots[i].Hora)

		for _, b := range bloqueios {
			if m >= b.ini && m < b.fim {
				slots[i].Disponivel = false
			}
		}

		for _, ag := range ocupados {
			ini := minutos(ag.Hora)
			fim := minutos(ag.HoraFim)
			if fim <= ini {
				fim = ini + ag.DuracaoTotal
			}

			if m >= ini && m < fim {
				idAg := ag.ID
				slots[i].Disponivel = false
				slots[i].Agendamento = &idAg
				slots[i].Servicos = ag.Servicos
			}
		}
	}

	return slots
}

func (s *Service) HorariosDia(ctx context.Context, req HorariosDiaRequest) (HorariosDiaResponse, error) {
	prof, err := s.profissionalDoSalao(ctx, req.BaseRequest)
	if err != nil {
		return HorariosDiaResponse{}, err
	}

	if strings.TrimSpace(req.Data) == "" {
		return HorariosDiaResponse{}, fmt.Errorf("%w: data obrigatória", ErrDadosInvalidos)
	}
	dia, err := parseData(req.Data)
	if err != nil {
		return HorariosDiaResponse{}, err
	}

	resposta := HorariosDiaResponse{UsaHorarios: prof.UsaHorarios, Horarios: []Slot{}}

	if prof.UsaHorarios == 0 {
		return resposta, nil
	}

	bloqueado, err := s.repository.DiaBloqueado(ctx, prof.ID, dia.Format(layoutData))
	if err != nil {
		return HorariosDiaResponse{}, err
	}
	if bloqueado {
		resposta.Bloqueado = true
		return resposta, nil
	}

	grade, err := s.repository.BuscarGrade(ctx, req.IDSalao, prof.ID, diaSemanaSigla(dia))
	if err != nil {
		return HorariosDiaResponse{}, err
	}
	if grade == nil || !grade.Abre {
		return resposta, nil
	}

	inicioDia := dia.Format(layoutSQL)
	fimDia := dia.AddDate(0, 0, 1).Format(layoutSQL)

	agendamentos, err := s.repository.ListarAgendamentos(ctx, FiltroAgendamentos{
		IDSalao:        req.IDSalao,
		ProfissionalID: prof.ID,
		Inicio:         inicioDia,
		Fim:            fimDia,
	})
	if err != nil {
		return HorariosDiaResponse{}, err
	}

	ocupados := make([]AgendamentoItem, 0, len(agendamentos))
	for _, ag := range agendamentos {
		if ag.Status != "cancelado" && ag.Status != "nao_compareceu" {
			ocupados = append(ocupados, ag)
		}
	}

	bloqueios, err := s.repository.BloqueiosDoDia(ctx, prof.ID, inicioDia, fimDia)
	if err != nil {
		return HorariosDiaResponse{}, err
	}

	slots := gerarSlots(grade, ocupados, bloqueiosEmMinutos(bloqueios, dia))

	// Se for hoje, esconde os horários que já passaram.
	agoraBR := agora()
	if dia.Format(layoutData) == agoraBR.Format(layoutData) {
		limite := agoraBR.Hour()*60 + agoraBR.Minute()
		futuros := make([]Slot, 0, len(slots))
		for _, sl := range slots {
			if minutos(sl.Hora) > limite {
				futuros = append(futuros, sl)
			}
		}
		slots = futuros
	}

	resposta.Horarios = slots
	return resposta, nil
}

// ============================================================================
// FATURAMENTO
// ============================================================================

func (s *Service) Faturamento(ctx context.Context, req FaturamentoRequest) (FaturamentoResponse, error) {
	prof, err := s.profissionalDoSalao(ctx, req.BaseRequest)
	if err != nil {
		return FaturamentoResponse{}, err
	}

	if req.Mes < 1 || req.Mes > 12 || req.Ano < 2000 || req.Ano > 2100 {
		return FaturamentoResponse{}, fmt.Errorf("%w: ano e mes obrigatórios e válidos", ErrDadosInvalidos)
	}

	inicio := time.Date(req.Ano, time.Month(req.Mes), 1, 0, 0, 0, 0, fusoBrasil)
	proximoMes := inicio.AddDate(0, 1, 0)
	ultimoDia := proximoMes.AddDate(0, 0, -1)

	linhas, err := s.repository.LinhasFaturamento(
		ctx, req.IDSalao, prof.ID,
		inicio.Format(layoutSQL), proximoMes.Format(layoutSQL),
	)
	if err != nil {
		return FaturamentoResponse{}, err
	}

	comissaoAplica := prof.TipoRemuneracao == "comissao" || prof.TipoRemuneracao == "misto"
	fixoAplica := prof.TipoRemuneracao == "fixo" || prof.TipoRemuneracao == "misto"

	var faturamento, comissao, fixo float64

	for _, l := range linhas {
		faturamento += l.PrecoTotal
		if comissaoAplica {
			comissao += l.PrecoTotal * l.PercentualComissao / 100
		}
	}

	// O salário fixo entra UMA vez no período (o Node somava por agendamento).
	if fixoAplica {
		fixo = prof.SalarioFixo
	}

	return FaturamentoResponse{
		Profissional: FaturamentoProfissional{
			Nome:                prof.Nome,
			TipoRemuneracao:     prof.TipoRemuneracao,
			FrequenciaPagamento: prof.FrequenciaPagamento,
			PercentualComissao:  prof.PercentualComissao,
			SalarioFixo:         prof.SalarioFixo,
		},
		Periodo: FaturamentoPeriodo{
			Ano:            req.Ano,
			Mes:            req.Mes,
			QuantidadeDias: ultimoDia.Day(),
			DataInicio:     inicio.Format(layoutData),
			DataFim:        ultimoDia.Format(layoutData),
		},
		Estatisticas: FaturamentoEstatisticas{
			TotalAtendimentos: len(linhas),
			FaturamentoTotal:  arredondar(faturamento),
			ValorComissao:     arredondar(comissao),
			ValorFixo:         arredondar(fixo),
			ValorAReceber:     arredondar(comissao + fixo),
		},
	}, nil
}