package painel

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// Limitador de tentativas erradas de código, por IP.
// Protege contra tentativa de adivinhar o código do salão.
type limitador struct {
	mu      sync.Mutex
	falhas  map[string]*registroFalha
	maximo  int
	janela  time.Duration
}

type registroFalha struct {
	quantidade int
	inicio     time.Time
}

func novoLimitador(maximo int, janela time.Duration) *limitador {
	return &limitador{
		falhas: make(map[string]*registroFalha),
		maximo: maximo,
		janela: janela,
	}
}

// bloqueado: já estourou o limite de erros dentro da janela?
func (l *limitador) bloqueado(chave string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	registro, ok := l.falhas[chave]
	if !ok {
		return false
	}

	if time.Since(registro.inicio) > l.janela {
		delete(l.falhas, chave)
		return false
	}

	return registro.quantidade >= l.maximo
}

func (l *limitador) registrarFalha(chave string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// limpeza ocasional para o mapa não crescer sem fim
	if len(l.falhas) > 10000 {
		for k, v := range l.falhas {
			if time.Since(v.inicio) > l.janela {
				delete(l.falhas, k)
			}
		}
	}

	registro, ok := l.falhas[chave]

	if !ok || time.Since(registro.inicio) > l.janela {
		l.falhas[chave] = &registroFalha{
			quantidade: 1,
			inicio:     time.Now(),
		}
		return
	}

	registro.quantidade++
}

func ipDe(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}