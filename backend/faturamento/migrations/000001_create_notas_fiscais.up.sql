CREATE SEQUENCE IF NOT EXISTS notas_fiscais_numero_seq
    START WITH 1
    INCREMENT BY 1;

CREATE TABLE IF NOT EXISTS notas_fiscais (
    id BIGSERIAL PRIMARY KEY,
    numero BIGINT NOT NULL UNIQUE
        DEFAULT nextval('notas_fiscais_numero_seq'),
    status VARCHAR(10) NOT NULL DEFAULT 'ABERTA',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),

    CONSTRAINT chk_notas_fiscais_status
        CHECK (status IN ('ABERTA', 'FECHADA'))
);

CREATE TABLE IF NOT EXISTS itens_nota_fiscal (
    id BIGSERIAL PRIMARY KEY,
    nota_fiscal_id BIGINT NOT NULL,
    produto_codigo VARCHAR(100) NOT NULL,
    quantidade INTEGER NOT NULL,

    CONSTRAINT fk_itens_nota_fiscal_nota
        FOREIGN KEY (nota_fiscal_id)
        REFERENCES notas_fiscais(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_itens_nota_fiscal_quantidade
        CHECK (quantidade > 0)
);

CREATE INDEX IF NOT EXISTS idx_itens_nota_fiscal_nota_fiscal_id
    ON itens_nota_fiscal (nota_fiscal_id);