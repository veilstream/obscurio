package psql_proxy

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDBManager_ProxyQuery(t *testing.T) {
	t.Skip("skipping test")
	queries := []string{
		`SELECT pg_catalog.set_config('search_path', '', false);`,
		`SELECT pg_catalog.pg_is_in_recovery()`,
		`SELECT pg_catalog.set_config('search_path', '', false);`,
		`SET DATESTYLE = ISO`,
		`SET INTERVALSTYLE = POSTGRES`,
		`SET extra_float_digits TO 3`,
		`SET synchronize_seqscans TO off`,
		`SET statement_timeout = 0`,
		`SET lock_timeout = 0`,
		`SET idle_in_transaction_session_timeout = 0`,
		`SET row_security = off`,
		`BEGIN`,
		`SET TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ ONLY`,
		`SELECT oid, rolname FROM pg_catalog.pg_roles ORDER BY 1`,
		`SELECT x.tableoid, x.oid, x.extname, n.nspname, x.extrelocatable, x.extversion, x.extconfig, x.extcondition FROM pg_extension x JOIN pg_namespace n ON n.oid = x.extnamespace`,
		`SELECT classid, objid, refobjid FROM pg_depend WHERE refclassid = 'pg_extension'::regclass AND deptype = 'e' ORDER BY 3`,
		`SELECT n.tableoid, n.oid, n.nspname, n.nspowner, n.nspacl, acldefault('n', n.nspowner) AS acldefault FROM pg_namespace n`,
		`SELECT c.tableoid, c.oid, c.relname, c.relnamespace, c.relkind, c.reltype, c.relowner, c.relchecks, c.relhasindex, c.relhasrules, c.relpages, c.relhastriggers, c.relpersistence, c.reloftype, c.relacl, acldefault(CASE WHEN c.relkind = 'S' THEN 's'::"char" ELSE 'r'::"char" END, c.relowner) AS acldefault, CASE WHEN c.relkind = 'f' THEN (SELECT ftserver FROM pg_catalog.pg_foreign_table WHERE ftrelid = c.oid) ELSE 0 END AS foreignserver, c.relfrozenxid, tc.relfrozenxid AS tfrozenxid, tc.oid AS toid, tc.relpages AS toastpages, tc.reloptions AS toast_reloptions, d.refobjid AS owning_tab, d.refobjsubid AS owning_col, tsp.spcname AS reltablespace, false AS relhasoids, c.relispopulated, c.relreplident, c.relrowsecurity, c.relforcerowsecurity, c.relminmxid, tc.relminmxid AS tminmxid, array_remove(array_remove(c.reloptions,'check_option=local'),'check_option=cascaded') AS reloptions, CASE WHEN 'check_option=local' = ANY (c.reloptions) THEN 'LOCAL'::text WHEN 'check_option=cascaded' = ANY (c.reloptions) THEN 'CASCADED'::text ELSE NULL END AS checkoption, am.amname, (d.deptype = 'i') IS TRUE AS is_identity_sequence, c.relispartition AS ispartition  FROM pg_class c LEFT JOIN pg_depend d ON (c.relkind = 'S' AND d.classid = 'pg_class'::regclass AND d.objid = c.oid AND d.objsubid = 0 AND d.refclassid = 'pg_class'::regclass AND d.deptype IN ('a', 'i')) LEFT JOIN pg_tablespace tsp ON (tsp.oid = c.reltablespace) LEFT JOIN pg_am am ON (c.relam = am.oid) LEFT JOIN pg_class tc ON (c.reltoastrelid = tc.oid AND tc.relkind = 't' AND c.relkind <> 'p') WHERE c.relkind IN ('r', 'S', 'v', 'c', 'm', 'f', 'p') ORDER BY c.oid`,
		`LOCK TABLE public.album, public.artist, public.customer, public.employee, public.genre, public.invoice, public.invoice_line, public.media_type, public.playlist, public.playlist_track, public.track IN ACCESS SHARE MODE`,
		`SELECT p.tableoid, p.oid, p.proname, p.prolang, p.pronargs, p.proargtypes, p.prorettype, p.proacl, acldefault('f', p.proowner) AS acldefault, p.pronamespace, p.proowner FROM pg_proc p LEFT JOIN pg_init_privs pip ON (p.oid = pip.objoid AND pip.classoid = 'pg_proc'::regclass AND pip.objsubid = 0) WHERE p.prokind <> 'a'   AND NOT EXISTS (SELECT 1 FROM pg_depend WHERE classid = 'pg_proc'::regclass AND objid = p.oid AND deptype = 'i')   AND (   pronamespace != (SELECT oid FROM pg_namespace WHERE nspname = 'pg_catalog')   OR EXISTS (SELECT 1 FROM pg_cast   WHERE pg_cast.oid > 16383    AND p.oid = pg_cast.castfunc)   OR EXISTS (SELECT 1 FROM pg_transform   WHERE pg_transform.oid > 16383 AND    (p.oid = pg_transform.trffromsql   OR p.oid = pg_transform.trftosql))   OR p.proacl IS DISTINCT FROM pip.initprivs)`,
		`SELECT tableoid, oid, typname, typnamespace, typacl, acldefault('T', typowner) AS acldefault, typowner, typelem, typrelid, CASE WHEN typrelid = 0 THEN ' '::"char" ELSE (SELECT relkind FROM pg_class WHERE oid = typrelid) END AS typrelkind, typtype, typisdefined, typname[0] = '_' AND typelem != 0 AND (SELECT typarray FROM pg_type te WHERE oid = pg_type.typelem) = oid AS isarray FROM pg_type`,
		`SELECT tableoid, oid, lanname, lanpltrusted, lanplcallfoid, laninline, lanvalidator, lanacl, acldefault('l', lanowner) AS acldefault, lanowner FROM pg_language WHERE lanispl ORDER BY oid`,
		`SELECT p.tableoid, p.oid, p.proname AS aggname, p.pronamespace AS aggnamespace, p.pronargs, p.proargtypes, p.proowner, p.proacl AS aggacl, acldefault('f', p.proowner) AS acldefault FROM pg_proc p LEFT JOIN pg_init_privs pip ON (p.oid = pip.objoid AND pip.classoid = 'pg_proc'::regclass AND pip.objsubid = 0) WHERE p.prokind = 'a' AND (p.pronamespace != (SELECT oid FROM pg_namespace WHERE nspname = 'pg_catalog') OR p.proacl IS DISTINCT FROM pip.initprivs)`,
		`SELECT tableoid, oid, oprname, oprnamespace, oprowner, oprkind, oprcode::oid AS oprcode FROM pg_operator`,
		`SELECT tableoid, oid, amname, amtype, amhandler::pg_catalog.regproc AS amhandler FROM pg_am`,
		`SELECT tableoid, oid, opcname, opcnamespace, opcowner FROM pg_opclass`,
		`SELECT tableoid, oid, opfname, opfnamespace, opfowner FROM pg_opfamily`,
		`SELECT tableoid, oid, prsname, prsnamespace, prsstart::oid, prstoken::oid, prsend::oid, prsheadline::oid, prslextype::oid FROM pg_ts_parser`,
		`SELECT tableoid, oid, tmplname, tmplnamespace, tmplinit::oid, tmpllexize::oid FROM pg_ts_template`,
		`SELECT tableoid, oid, dictname, dictnamespace, dictowner, dicttemplate, dictinitoption FROM pg_ts_dict`,
		`SELECT tableoid, oid, cfgname, cfgnamespace, cfgowner, cfgparser FROM pg_ts_config`,
		`SELECT tableoid, oid, fdwname, fdwowner, fdwhandler::pg_catalog.regproc, fdwvalidator::pg_catalog.regproc, fdwacl, acldefault('F', fdwowner) AS acldefault, array_to_string(ARRAY(SELECT quote_ident(option_name) || ' ' || quote_literal(option_value) FROM pg_options_to_table(fdwoptions) ORDER BY option_name), E',     ') AS fdwoptions FROM pg_foreign_data_wrapper`,
		`SELECT tableoid, oid, srvname, srvowner, srvfdw, srvtype, srvversion, srvacl, acldefault('S', srvowner) AS acldefault, array_to_string(ARRAY(SELECT quote_ident(option_name) || ' ' || quote_literal(option_value) FROM pg_options_to_table(srvoptions) ORDER BY option_name), E',     ') AS srvoptions FROM pg_foreign_server`,
		`SELECT oid, tableoid, defaclrole, defaclnamespace, defaclobjtype, defaclacl, CASE WHEN defaclnamespace = 0 THEN acldefault(CASE WHEN defaclobjtype = 'S' THEN 's'::"char" ELSE defaclobjtype END, defaclrole) ELSE '{}' END AS acldefault FROM pg_default_acl`,
		`SELECT tableoid, oid, collname, collnamespace, collowner FROM pg_collation`,
		`SELECT tableoid, oid, conname, connamespace, conowner FROM pg_conversion`,
		`SELECT tableoid, oid, castsource, casttarget, castfunc, castcontext, castmethod FROM pg_cast c WHERE NOT EXISTS ( SELECT 1 FROM pg_range r WHERE c.castsource = r.rngtypid AND c.casttarget = r.rngmultitypid ) ORDER BY 3,4`,
		`SELECT tableoid, oid, trftype, trflang, trffromsql::oid, trftosql::oid FROM pg_transform ORDER BY 3,4`,
		`SELECT inhrelid, inhparent FROM pg_inherits`,
		`SELECT e.tableoid, e.oid, evtname, evtenabled, evtevent, evtowner, array_to_string(array(select quote_literal(x)  from unnest(evttags) as t(x)), ', ') as evttags, e.evtfoid::regproc as evtfname FROM pg_event_trigger e ORDER BY e.oid`,
		`SELECT conrelid, confrelid FROM pg_constraint JOIN pg_depend ON (objid = confrelid) WHERE contype = 'f' AND refclassid = 'pg_extension'::regclass AND classid = 'pg_class'::regclass;`,
		`SELECT a.attrelid, a.attnum, a.attname, a.attstattarget, a.attstorage, t.typstorage, a.attnotnull, a.atthasdef, a.attisdropped, a.attlen, a.attalign, a.attislocal, pg_catalog.format_type(t.oid, a.atttypmod) AS atttypname, array_to_string(a.attoptions, ', ') AS attoptions, CASE WHEN a.attcollation <> t.typcollation THEN a.attcollation ELSE 0 END AS attcollation, pg_catalog.array_to_string(ARRAY(SELECT pg_catalog.quote_ident(option_name) || ' ' || pg_catalog.quote_literal(option_value) FROM pg_catalog.pg_options_to_table(attfdwoptions) ORDER BY option_name), E',     ') AS attfdwoptions, a.attcompression AS attcompression, a.attidentity, CASE WHEN a.atthasmissing AND NOT a.attisdropped THEN a.attmissingval ELSE null END AS attmissingval, a.attgenerated FROM unnest('{16660,16665,16670,16675,16680,16685,16690,16695,16700,16705,16710}'::pg_catalog.oid[]) AS src(tbloid) JOIN pg_catalog.pg_attribute a ON (src.tbloid = a.attrelid) LEFT JOIN pg_catalog.pg_type t ON (a.atttypid = t.oid) WHERE a.attnum > 0::pg_catalog.int2 ORDER BY a.attrelid, a.attnum`,
		`SELECT partrelid FROM pg_partitioned_table WHERE (SELECT c.oid FROM pg_opclass c JOIN pg_am a ON c.opcmethod = a.oid WHERE opcname = 'enum_ops' AND opcnamespace = 'pg_catalog'::regnamespace AND amname = 'hash') = ANY(partclass)`,
		`SELECT t.tableoid, t.oid, i.indrelid, t.relname AS indexname, pg_catalog.pg_get_indexdef(i.indexrelid) AS indexdef, i.indkey, i.indisclustered, c.contype, c.conname, c.condeferrable, c.condeferred, c.tableoid AS contableoid, c.oid AS conoid, pg_catalog.pg_get_constraintdef(c.oid, false) AS condef, (SELECT spcname FROM pg_catalog.pg_tablespace s WHERE s.oid = t.reltablespace) AS tablespace, t.reloptions AS indreloptions, i.indisreplident, inh.inhparent AS parentidx, i.indnkeyatts AS indnkeyatts, i.indnatts AS indnatts, (SELECT pg_catalog.array_agg(attnum ORDER BY attnum)   FROM pg_catalog.pg_attribute   WHERE attrelid = i.indexrelid AND     attstattarget >= 0) AS indstatcols, (SELECT pg_catalog.array_agg(attstattarget ORDER BY attnum)   FROM pg_catalog.pg_attribute   WHERE attrelid = i.indexrelid AND     attstattarget >= 0) AS indstatvals, i.indnullsnotdistinct FROM unnest('{16660,16665,16670,16675,16680,16685,16690,16695,16700,16705,16710}'::pg_catalog.oid[]) AS src(tbloid) JOIN pg_catalog.pg_index i ON (src.tbloid = i.indrelid) JOIN pg_catalog.pg_class t ON (t.oid = i.indexrelid) JOIN pg_catalog.pg_class t2 ON (t2.oid = i.indrelid) LEFT JOIN pg_catalog.pg_constraint c ON (i.indrelid = c.conrelid AND i.indexrelid = c.conindid AND c.contype IN ('p','u','x')) LEFT JOIN pg_catalog.pg_inherits inh ON (inh.inhrelid = indexrelid) WHERE (i.indisvalid OR t2.relkind = 'p') AND i.indisready ORDER BY i.indrelid, indexname`,
		`SELECT tableoid, oid, stxname, stxnamespace, stxowner, stxrelid, stxstattarget FROM pg_catalog.pg_statistic_ext`,
		`SELECT c.tableoid, c.oid, conrelid, conname, confrelid, conindid, pg_catalog.pg_get_constraintdef(c.oid) AS condef FROM unnest('{16660,16665,16670,16675,16680,16685,16690,16695,16700,16705,16710}'::pg_catalog.oid[]) AS src(tbloid) JOIN pg_catalog.pg_constraint c ON (src.tbloid = c.conrelid) WHERE contype = 'f' AND conparentid = 0 ORDER BY conrelid, conname`,
		`SELECT t.tgrelid, t.tgname, t.tgfoid::pg_catalog.regproc AS tgfname, pg_catalog.pg_get_triggerdef(t.oid, false) AS tgdef, t.tgenabled, t.tableoid, t.oid, t.tgparentid <> 0 AS tgispartition FROM unnest('{16660,16665,16670,16675,16680,16685,16690,16695,16700,16705,16710}'::pg_catalog.oid[]) AS src(tbloid) JOIN pg_catalog.pg_trigger t ON (src.tbloid = t.tgrelid) LEFT JOIN pg_catalog.pg_trigger u ON (u.oid = t.tgparentid) WHERE ((NOT t.tgisinternal AND t.tgparentid = 0) OR t.tgenabled != u.tgenabled) ORDER BY t.tgrelid, t.tgname`,
		`SELECT tableoid, oid, rulename, ev_class AS ruletable, ev_type, is_instead, ev_enabled FROM pg_rewrite ORDER BY oid`,
		`SELECT pol.oid, pol.tableoid, pol.polrelid, pol.polname, pol.polcmd, pol.polpermissive, CASE WHEN pol.polroles = '{0}' THEN NULL ELSE    pg_catalog.array_to_string(ARRAY(SELECT pg_catalog.quote_ident(rolname) from pg_catalog.pg_roles WHERE oid = ANY(pol.polroles)), ', ') END AS polroles, pg_catalog.pg_get_expr(pol.polqual, pol.polrelid) AS polqual, pg_catalog.pg_get_expr(pol.polwithcheck, pol.polrelid) AS polwithcheck FROM unnest('{16660,16665,16670,16675,16680,16685,16690,16695,16700,16705,16710}'::pg_catalog.oid[]) AS src(tbloid) JOIN pg_catalog.pg_policy pol ON (src.tbloid = pol.polrelid)`,
		`SELECT p.tableoid, p.oid, p.pubname, p.pubowner, p.puballtables, p.pubinsert, p.pubupdate, p.pubdelete, p.pubtruncate, p.pubviaroot FROM pg_publication p`,
		`SELECT tableoid, oid, prpubid, prrelid, pg_catalog.pg_get_expr(prqual, prrelid) AS prrelqual, (CASE   WHEN pr.prattrs IS NOT NULL THEN     (SELECT array_agg(attname)        FROM          pg_catalog.generate_series(0, pg_catalog.array_upper(pr.prattrs::pg_catalog.int2[], 1)) s,          pg_catalog.pg_attribute       WHERE attrelid = pr.prrelid AND attnum = prattrs[s])   ELSE NULL END) prattrs FROM pg_catalog.pg_publication_rel pr`,
		`SELECT tableoid, oid, pnpubid, pnnspid FROM pg_catalog.pg_publication_namespace`,
		`SELECT count(*) FROM pg_subscription WHERE subdbid = (SELECT oid FROM pg_database                 WHERE datname = current_database())`,
		`WITH RECURSIVE w AS ( SELECT d1.objid, d2.refobjid, c2.relkind AS refrelkind FROM pg_depend d1 JOIN pg_class c1 ON c1.oid = d1.objid AND c1.relkind = 'm' JOIN pg_rewrite r1 ON r1.ev_class = d1.objid JOIN pg_depend d2 ON d2.classid = 'pg_rewrite'::regclass AND d2.objid = r1.oid AND d2.refobjid <> d1.objid JOIN pg_class c2 ON c2.oid = d2.refobjid AND c2.relkind IN ('m','v') WHERE d1.classid = 'pg_class'::regclass UNION SELECT w.objid, d3.refobjid, c3.relkind FROM w JOIN pg_rewrite r3 ON r3.ev_class = w.refobjid JOIN pg_depend d3 ON d3.classid = 'pg_rewrite'::regclass AND d3.objid = r3.oid AND d3.refobjid <> w.refobjid JOIN pg_class c3 ON c3.oid = d3.refobjid AND c3.relkind IN ('m','v') ) SELECT 'pg_class'::regclass::oid AS classid, objid, refobjid FROM w WHERE refrelkind = 'm'`,
		`SELECT oid, lomowner, lomacl, acldefault('L', lomowner) AS acldefault FROM pg_largeobject_metadata`,
		`SELECT classid, objid, refclassid, refobjid, deptype FROM pg_depend WHERE deptype != 'p' AND deptype != 'e' UNION ALL SELECT 'pg_opfamily'::regclass AS classid, amopfamily AS objid, refclassid, refobjid, deptype FROM pg_depend d, pg_amop o WHERE deptype NOT IN ('p', 'e', 'i') AND classid = 'pg_amop'::regclass AND objid = o.oid AND NOT (refclassid = 'pg_opfamily'::regclass AND amopfamily = refobjid) UNION ALL SELECT 'pg_opfamily'::regclass AS classid, amprocfamily AS objid, refclassid, refobjid, deptype FROM pg_depend d, pg_amproc p WHERE deptype NOT IN ('p', 'e', 'i') AND classid = 'pg_amproc'::regclass AND objid = p.oid AND NOT (refclassid = 'pg_opfamily'::regclass AND amprocfamily = refobjid) ORDER BY 1,2`,
		`SELECT DISTINCT attrelid FROM pg_attribute WHERE attacl IS NOT NULL`,
		`SELECT objoid, classoid, objsubid, privtype, initprivs FROM pg_init_privs`,
		`SELECT description, classoid, objoid, objsubid FROM pg_catalog.pg_description ORDER BY classoid, objoid, objsubid`,
		`SELECT label, provider, classoid, objoid, objsubid FROM pg_catalog.pg_seclabel ORDER BY classoid, objoid, objsubid`,
		`SELECT pg_catalog.current_schemas(false)`,
		`PREPARE getColumnACLs(pg_catalog.oid) AS SELECT at.attname, at.attacl, '{}' AS acldefault, pip.privtype, pip.initprivs FROM pg_catalog.pg_attribute at LEFT JOIN pg_catalog.pg_init_privs pip ON (at.attrelid = pip.objoid AND pip.classoid = 'pg_catalog.pg_class'::pg_catalog.regclass AND at.attnum = pip.objsubid) WHERE at.attrelid = $1 AND NOT at.attisdropped AND (at.attacl IS NOT NULL OR pip.initprivs IS NOT NULL) ORDER BY at.attnum`,
		`EXECUTE getColumnACLs('6100')`,
		`COPY public.album (album_id, title, artist_id) TO stdout;`,
		`COPY public.artist (artist_id, name) TO stdout;`,
		`COPY public.customer (customer_id, first_name, last_name, company, address, city, state, country, postal_code, phone, fax, email, support_rep_id) TO stdout;`,
		`COPY public.employee (employee_id, last_name, first_name, title, reports_to, birth_date, hire_date, address, city, state, country, postal_code, phone, fax, email) TO stdout;`,
		`COPY public.genre (genre_id, name) TO stdout;`,
		`COPY public.invoice (invoice_id, customer_id, invoice_date, billing_address, billing_city, billing_state, billing_country, billing_postal_code, total) TO stdout;`,
		`COPY public.invoice_line (invoice_line_id, invoice_id, track_id, unit_price, quantity) TO stdout;`,
		`COPY public.media_type (media_type_id, name) TO stdout;`,
		`COPY public.playlist (playlist_id, name) TO stdout;`,
		`COPY public.playlist_track (playlist_id, track_id) TO stdout;`,
		`COPY public.track (track_id, name, album_id, media_type_id, genre_id, composer, milliseconds, bytes, unit_price) TO stdout;`,
	}

	for n, query := range queries {
		t.Run(fmt.Sprintf("query[%d]", n), func(t *testing.T) {
			conf, err := GetConfig()
			require.NoError(t, err)
			server, err := NewServer(conf, nil)
			require.NoError(t, err)
			ctx := context.Background()
			context.WithValue(ctx, 1, map[string]string{
				"user": "user",
			})
			sp, err := server.getDBManager(ctx)
			require.NoError(t, err)
			err = sp.ProxyQuery(query, nil)
			require.NoError(t, err)
		})
	}
}
