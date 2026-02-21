package entity

type PublicacionEntity struct {
	post_id               int64   `db:"post_id"`
	post_uuid             string  `db:"post_uuid"`
	organizacion_id       int64   `db:"organizacion_id"`
	titulo                string  `db:"titulo"`
	descripcion           string  `db:"descripcion"`
	categoria_id          *int64  `db:"categoria_id"`       // Puede ser nulo
	multimedia_data       string  `db:"multimedia_data"`    // JSONB como string
	cliente_externo_id    *int64  `db:"cliente_externo_id"` // Puede ser nulo
	verificado            bool    `db:"verificado"`
	count_likes           int     `db:"count_likes"`
	count_comments        int     `db:"count_comments"`
	count_shares          int     `db:"count_shares"`
	promedio_calificacion float64 `db:"promedio_calificacion"`
	fecha_posteo          string  `db:"fecha_posteo"` // TIMESTAMPTZ como string
}
