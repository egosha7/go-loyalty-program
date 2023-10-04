--
-- PostgreSQL database dump
--

-- Dumped from database version 16.0
-- Dumped by pg_dump version 16.0

-- Started on 2023-10-04 00:12:52

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 217 (class 1259 OID 16414)
-- Name: loyalty_balance; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.loyalty_balance (
    loyalty_id integer NOT NULL,
    user_id integer,
    points integer
);


ALTER TABLE public.loyalty_balance OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 16419)
-- Name: loyalty_withdrawals; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.loyalty_withdrawals (
    withdrawal_id integer NOT NULL,
    user_id integer,
    order_id integer,
    withdrawn_points integer
);


ALTER TABLE public.loyalty_withdrawals OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 16471)
-- Name: loyalty_withdrawals_withdrawal_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.loyalty_withdrawals ALTER COLUMN withdrawal_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.loyalty_withdrawals_withdrawal_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 216 (class 1259 OID 16404)
-- Name: orders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.orders (
    order_id integer NOT NULL,
    user_id integer,
    order_status text,
    order_number text,
    "timestamp" timestamp with time zone
);


ALTER TABLE public.orders OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 16454)
-- Name: orders_order_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.orders ALTER COLUMN order_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.orders_order_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 215 (class 1259 OID 16399)
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    user_id integer NOT NULL,
    login text,
    password text
);


ALTER TABLE public.users OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 16453)
-- Name: users_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.users ALTER COLUMN user_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.users_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 4707 (class 2606 OID 16418)
-- Name: loyalty_balance loyalty_balance_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.loyalty_balance
    ADD CONSTRAINT loyalty_balance_pkey PRIMARY KEY (loyalty_id);


--
-- TOC entry 4709 (class 2606 OID 16423)
-- Name: loyalty_withdrawals loyalty_withdrawals_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.loyalty_withdrawals
    ADD CONSTRAINT loyalty_withdrawals_pkey PRIMARY KEY (withdrawal_id);


--
-- TOC entry 4705 (class 2606 OID 16408)
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (order_id);


--
-- TOC entry 4703 (class 2606 OID 16403)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- TOC entry 4711 (class 2606 OID 16434)
-- Name: loyalty_balance loyalty_balance_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.loyalty_balance
    ADD CONSTRAINT loyalty_balance_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) NOT VALID;


--
-- TOC entry 4712 (class 2606 OID 16429)
-- Name: loyalty_withdrawals loyalty_withdrawals_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.loyalty_withdrawals
    ADD CONSTRAINT loyalty_withdrawals_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(order_id);


--
-- TOC entry 4713 (class 2606 OID 16424)
-- Name: loyalty_withdrawals loyalty_withdrawals_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.loyalty_withdrawals
    ADD CONSTRAINT loyalty_withdrawals_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- TOC entry 4710 (class 2606 OID 16409)
-- Name: orders users_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT users_fk FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON UPDATE CASCADE ON DELETE CASCADE;


-- Completed on 2023-10-04 00:12:52

--
-- PostgreSQL database dump complete
--

