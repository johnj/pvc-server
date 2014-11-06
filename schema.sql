--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: hosts; Type: TABLE; Schema: public; Owner: pvc; Tablespace: 
--

CREATE TABLE hosts (
    certname character varying(4096) NOT NULL,
    force_run integer DEFAULT 0,
    get_facts integer DEFAULT 0,
    monitored_files text,
    updated_at timestamp with time zone,
    created_at timestamp with time zone,
    check_interval integer DEFAULT 60,
    files_monitored character varying(8192) DEFAULT ''::character varying,
    puppetvars text,
    cpus integer,
    max_runinterval integer,
    last_run_started timestamp with time zone,
    last_run_finished timestamp with time zone
);


ALTER TABLE public.hosts OWNER TO pvc;

--
-- Name: ppms; Type: TABLE; Schema: public; Owner: pvc; Tablespace: 
--

CREATE TABLE ppms (
    fqdn character varying(1024),
    updated_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    active_processes integer,
    waiting_processes integer,
    loadavg integer,
    weighting integer DEFAULT 1,
    allowed_processes integer DEFAULT 0,
    processed_requests bigint DEFAULT 0,
    score integer,
    cpus integer
);


ALTER TABLE public.ppms OWNER TO pvc;

--
-- Name: hosts_pkey; Type: CONSTRAINT; Schema: public; Owner: pvc; Tablespace: 
--

ALTER TABLE ONLY hosts
    ADD CONSTRAINT hosts_pkey PRIMARY KEY (certname);


--
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- Name: ppms; Type: ACL; Schema: public; Owner: pvc
--

REVOKE ALL ON TABLE ppms FROM PUBLIC;
REVOKE ALL ON TABLE ppms FROM pvc;
GRANT ALL ON TABLE ppms TO pvc;
GRANT ALL ON TABLE ppms TO postgres;


--
-- PostgreSQL database dump complete
--

