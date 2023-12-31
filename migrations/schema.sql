--
-- PostgreSQL database dump
--

-- Dumped from database version 15.5
-- Dumped by pg_dump version 15.5 (Ubuntu 15.5-1.pgdg22.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: jofosuware
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO jofosuware;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: customers; Type: TABLE; Schema: public; Owner: jofosuware
--

CREATE TABLE public.customers (
    id integer NOT NULL,
    customer_id character varying,
    cust_image bytea,
    id_type character varying,
    card_image bytea,
    first_name character varying,
    last_name character varying,
    house_address character varying,
    phone integer,
    location character varying,
    landmark character varying,
    agreement character varying,
    contract_status character varying,
    months integer,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


ALTER TABLE public.customers OWNER TO jofosuware;

--
-- Name: customers_id_seq; Type: SEQUENCE; Schema: public; Owner: jofosuware
--

CREATE SEQUENCE public.customers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.customers_id_seq OWNER TO jofosuware;

--
-- Name: customers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: jofosuware
--

ALTER SEQUENCE public.customers_id_seq OWNED BY public.customers.id;


--
-- Name: payments; Type: TABLE; Schema: public; Owner: jofosuware
--

CREATE TABLE public.payments (
    id integer NOT NULL,
    customer_id character varying,
    month character varying,
    amount real,
    payment_date timestamp without time zone,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


ALTER TABLE public.payments OWNER TO jofosuware;

--
-- Name: payments_id_seq; Type: SEQUENCE; Schema: public; Owner: jofosuware
--

CREATE SEQUENCE public.payments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.payments_id_seq OWNER TO jofosuware;

--
-- Name: payments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: jofosuware
--

ALTER SEQUENCE public.payments_id_seq OWNED BY public.payments.id;


--
-- Name: products; Type: TABLE; Schema: public; Owner: jofosuware
--

CREATE TABLE public.products (
    id integer NOT NULL,
    serial character varying,
    name character varying,
    description text,
    price real,
    units integer,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


ALTER TABLE public.products OWNER TO jofosuware;

--
-- Name: products_id_seq; Type: SEQUENCE; Schema: public; Owner: jofosuware
--

CREATE SEQUENCE public.products_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.products_id_seq OWNER TO jofosuware;

--
-- Name: products_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: jofosuware
--

ALTER SEQUENCE public.products_id_seq OWNED BY public.products.id;


--
-- Name: purchased_oncredit; Type: TABLE; Schema: public; Owner: jofosuware
--

CREATE TABLE public.purchased_oncredit (
    id integer NOT NULL,
    customer_id character varying,
    serial character varying,
    price real,
    quantity integer,
    deposit real,
    balance real,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


ALTER TABLE public.purchased_oncredit OWNER TO jofosuware;

--
-- Name: purchased_oncredit_id_seq; Type: SEQUENCE; Schema: public; Owner: jofosuware
--

CREATE SEQUENCE public.purchased_oncredit_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.purchased_oncredit_id_seq OWNER TO jofosuware;

--
-- Name: purchased_oncredit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: jofosuware
--

ALTER SEQUENCE public.purchased_oncredit_id_seq OWNED BY public.purchased_oncredit.id;


--
-- Name: purchases; Type: TABLE; Schema: public; Owner: jofosuware
--

CREATE TABLE public.purchases (
    id integer NOT NULL,
    serial character varying,
    quantity integer,
    amount real,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


ALTER TABLE public.purchases OWNER TO jofosuware;

--
-- Name: purchases_id_seq; Type: SEQUENCE; Schema: public; Owner: jofosuware
--

CREATE SEQUENCE public.purchases_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.purchases_id_seq OWNER TO jofosuware;

--
-- Name: purchases_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: jofosuware
--

ALTER SEQUENCE public.purchases_id_seq OWNED BY public.purchases.id;


--
-- Name: schema_migration; Type: TABLE; Schema: public; Owner: jofosuware
--

CREATE TABLE public.schema_migration (
    version character varying(14) NOT NULL
);


ALTER TABLE public.schema_migration OWNER TO jofosuware;

--
-- Name: users; Type: TABLE; Schema: public; Owner: jofosuware
--

CREATE TABLE public.users (
    id integer NOT NULL,
    first_name character varying,
    last_name character varying,
    user_name character varying,
    password character varying,
    user_image bytea,
    access_level character varying,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


ALTER TABLE public.users OWNER TO jofosuware;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: jofosuware
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO jofosuware;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: jofosuware
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: witness; Type: TABLE; Schema: public; Owner: jofosuware
--

CREATE TABLE public.witness (
    id integer NOT NULL,
    customer_id character varying,
    first_name character varying,
    last_name character varying,
    phone integer,
    terms character varying,
    witness_image bytea,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


ALTER TABLE public.witness OWNER TO jofosuware;

--
-- Name: witness_id_seq; Type: SEQUENCE; Schema: public; Owner: jofosuware
--

CREATE SEQUENCE public.witness_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.witness_id_seq OWNER TO jofosuware;

--
-- Name: witness_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: jofosuware
--

ALTER SEQUENCE public.witness_id_seq OWNED BY public.witness.id;


--
-- Name: customers id; Type: DEFAULT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.customers ALTER COLUMN id SET DEFAULT nextval('public.customers_id_seq'::regclass);


--
-- Name: payments id; Type: DEFAULT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.payments ALTER COLUMN id SET DEFAULT nextval('public.payments_id_seq'::regclass);


--
-- Name: products id; Type: DEFAULT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.products ALTER COLUMN id SET DEFAULT nextval('public.products_id_seq'::regclass);


--
-- Name: purchased_oncredit id; Type: DEFAULT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.purchased_oncredit ALTER COLUMN id SET DEFAULT nextval('public.purchased_oncredit_id_seq'::regclass);


--
-- Name: purchases id; Type: DEFAULT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.purchases ALTER COLUMN id SET DEFAULT nextval('public.purchases_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: witness id; Type: DEFAULT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.witness ALTER COLUMN id SET DEFAULT nextval('public.witness_id_seq'::regclass);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: payments payments_pkey; Type: CONSTRAINT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: purchased_oncredit purchased_oncredit_pkey; Type: CONSTRAINT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.purchased_oncredit
    ADD CONSTRAINT purchased_oncredit_pkey PRIMARY KEY (id);


--
-- Name: purchases purchases_pkey; Type: CONSTRAINT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.purchases
    ADD CONSTRAINT purchases_pkey PRIMARY KEY (id);


--
-- Name: schema_migration schema_migration_pkey; Type: CONSTRAINT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.schema_migration
    ADD CONSTRAINT schema_migration_pkey PRIMARY KEY (version);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: witness witness_pkey; Type: CONSTRAINT; Schema: public; Owner: jofosuware
--

ALTER TABLE ONLY public.witness
    ADD CONSTRAINT witness_pkey PRIMARY KEY (id);


--
-- Name: schema_migration_version_idx; Type: INDEX; Schema: public; Owner: jofosuware
--

CREATE UNIQUE INDEX schema_migration_version_idx ON public.schema_migration USING btree (version);


--
-- Name: TABLE pg_stat_database; Type: ACL; Schema: pg_catalog; Owner: postgres
--

GRANT SELECT ON TABLE pg_catalog.pg_stat_database TO datadog;


--
-- Name: DEFAULT PRIVILEGES FOR SEQUENCES; Type: DEFAULT ACL; Schema: -; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres GRANT ALL ON SEQUENCES  TO jofosuware;


--
-- Name: DEFAULT PRIVILEGES FOR TYPES; Type: DEFAULT ACL; Schema: -; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres GRANT ALL ON TYPES  TO jofosuware;


--
-- Name: DEFAULT PRIVILEGES FOR FUNCTIONS; Type: DEFAULT ACL; Schema: -; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres GRANT ALL ON FUNCTIONS  TO jofosuware;


--
-- Name: DEFAULT PRIVILEGES FOR TABLES; Type: DEFAULT ACL; Schema: -; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres GRANT ALL ON TABLES  TO jofosuware;


--
-- PostgreSQL database dump complete
--

