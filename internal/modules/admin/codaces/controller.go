package codaces

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
}

// ------------------------------------------------------------
// BUSCAR CÓDIGO
// ------------------------------------------------------------

func (c *Controller) BuscarCodigo(
	salaoID uint64,
) (*CodigoResponse, error) {

	return c.service.BuscarCodigo(salaoID)
}

// ------------------------------------------------------------
// INSERIR CÓDIGO
// ------------------------------------------------------------

func (c *Controller) InserirCodigo(
	salaoID uint64,
	codigo string,
) error {

	return c.service.InserirCodigo(salaoID, codigo)
}

// ------------------------------------------------------------
// ALTERAR CÓDIGO
// ------------------------------------------------------------

func (c *Controller) AlterarCodigo(
	salaoID uint64,
	codigo string,
) error {

	return c.service.AlterarCodigo(salaoID, codigo)
}