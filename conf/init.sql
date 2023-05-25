--  This file is part of the eliona project.
--  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
--  ______ _ _
-- |  ____| (_)
-- | |__  | |_  ___  _ __   __ _
-- |  __| | | |/ _ \| '_ \ / _` |
-- | |____| | | (_) | | | | (_| |
-- |______|_|_|\___/|_| |_|\__,_|
--
--  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
--  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
--  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
--  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
--  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

create schema if not exists kontakt_io;

-- Should be editable by eliona frontend.
create table if not exists kontakt_io.configuration
(
	id               bigserial primary key,
	api_key          text,
	absolute_x       integer not null default 0,
	absolute_y       integer not null default 0,
	refresh_interval integer not null default 60,
	request_timeout  integer not null default 120,
	asset_filter     json,
	active           boolean default false,
	enable           boolean default false,
	project_ids      text[]
);

-- Location corresponds to one asset in Eliona
-- Should be read-only by eliona frontend.
create table if not exists kontakt_io.location
(
	id               bigserial primary key,
	parent_id        bigserial references kontakt_io.location(id),
	configuration_id bigserial not null references kontakt_io.configuration(id),
	project_id       text      not null,
	global_asset_id  text      not null,
	floor_height     float,
	asset_id         integer
);

-- Tag corresponds to one asset in Eliona
-- Should be read-only by eliona frontend.
create table if not exists kontakt_io.tag
(
	configuration_id bigserial references kontakt_io.configuration(id),
	project_id       text      not null,
	global_asset_id  text      not null,
	asset_id         integer,
	primary key (configuration_id, project_id, global_asset_id)
);

-- Makes the new objects available for all other init steps
commit;
