/*
 Navicat Premium Data Transfer

 Source Server         : InvoiceD
 Source Server Type    : SQLite
 Source Server Version : 3007015
 Source Database       : main

 Target Server Type    : SQLite
 Target Server Version : 3007015
 File Encoding         : utf-8

 Date: 05/25/2016 19:59:47 PM
*/

PRAGMA foreign_keys = false;

-- ----------------------------
--  Table structure for "config"
-- ----------------------------
DROP TABLE IF EXISTS "config";
CREATE TABLE "config" (
	 "key" text NOT NULL,
	 "value_text" text,
	 "value_number" integer,
	PRIMARY KEY("key")
);

-- ----------------------------
--  Records of "main"."config"
-- ----------------------------
BEGIN;
INSERT INTO "config" VALUES ('version', null, 1);
COMMIT;

-- ----------------------------
--  Table structure for "customer"
-- ----------------------------
DROP TABLE IF EXISTS "customer";
CREATE TABLE "customer" (
	 "id" integer NOT NULL,
	 "name" text NOT NULL,
	 "address_one" text,
	 "address_two" text,
	 "country" text,
	PRIMARY KEY("id")
);

-- ----------------------------
--  Records of "main"."customer"
-- ----------------------------
BEGIN;
INSERT INTO "customer" VALUES (1, 'XS News B.V.', 'New Yorkstraat 9-13', '1175 RD Lijnden', 'NL');
COMMIT;

-- ----------------------------
--  Table structure for "entity"
-- ----------------------------
DROP TABLE IF EXISTS "entity";
CREATE TABLE "entity" (
	 "id" integer NOT NULL,
	 "name" text NOT NULL,
	 "bank_iban" text NOT NULL,
	 "bank_bic" text NOT NULL,
	 "number_vat" text,
	 "number_coc" text,
	 "owner_name" text NOT NULL,
	 "owner_address_one" text NOT NULL,
	 "owner_address_two" text,
	 "owner_country" text,
	PRIMARY KEY("id")
);

-- ----------------------------
--  Records of "main"."entity"
-- ----------------------------
BEGIN;
INSERT INTO "entity" VALUES (1, 'RootDev', 'NL12345', 'BIC', 'NL12345', 54321, 'M.P. Droog', 'Dorpsstraat 236A', '1713HP, Obdam', null);
COMMIT;

-- ----------------------------
--  Table structure for "invoice"
-- ----------------------------
DROP TABLE IF EXISTS "invoice";
CREATE TABLE "invoice" (
	 "id" integer NOT NULL,
	 "name" text NOT NULL,
	 "time_added" text NOT NULL,
	 "time_finalized" text,
	 "time_sent" text,
	 "time_paid" text,
	PRIMARY KEY("id")
);

-- ----------------------------
--  Table structure for "invoiceline"
-- ----------------------------
DROP TABLE IF EXISTS "invoiceline";
CREATE TABLE "invoiceline" (
	 "id" integer NOT NULL,
	 "invoice_id" integer,
	 "entity_id" integer NOT NULL,
	 "customer_id" integer NOT NULL,
	 "description" text NOT NULL,
	 "unit_quantity" integer NOT NULL,
	 "unit_price" real NOT NULL,
	 "total" real NOT NULL,
	 "vat" integer NOT NULL,
	 "followup_id" integer,
	 "ledger" text,
	PRIMARY KEY("id")
);

-- ----------------------------
--  Table structure for "ledger"
-- ----------------------------
DROP TABLE IF EXISTS "ledger";
CREATE TABLE "ledger" (
	 "code" integer NOT NULL,
	 "entity_id" integer NOT NULL,
	 "name" text NOT NULL,
	PRIMARY KEY("code")
);

-- ----------------------------
--  Table structure for "product"
-- ----------------------------
DROP TABLE IF EXISTS "product";
CREATE TABLE "product" (
	 "id" integer NOT NULL,
	 "description" text NOT NULL,
	 "date_start" text NOT NULL,
	 "date_end" text,
	 "unit_quantity" integer NOT NULL,
	 "unit_price" real NOT NULL,
	 "ledger" text,
	PRIMARY KEY("id")
);

-- ----------------------------
--  Indexes structure for table "customer"
-- ----------------------------
CREATE UNIQUE INDEX "unique_name" ON customer ("name" ASC);

-- ----------------------------
--  Indexes structure for table "entity"
-- ----------------------------
CREATE UNIQUE INDEX "unique_entity_name" ON entity ("name" ASC);
CREATE UNIQUE INDEX "unique_entity_vat" ON entity ("number_vat" ASC);
CREATE UNIQUE INDEX "unique_entity_coc" ON entity ("number_coc" ASC);

-- ----------------------------
--  Indexes structure for table "invoiceline"
-- ----------------------------
CREATE UNIQUE INDEX "unique_description" ON invoiceline ("description" ASC);

PRAGMA foreign_keys = true;
