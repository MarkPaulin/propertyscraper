CREATE TABLE properties (
	url text not null,
	id integer not null,
	updated_time text,
	lat float,
	lon float,
	address text,
	postcode text,
	price_offers text,
	price text,
	price_min text,
	price_max text,
	agent text,
	branch text,
	style text,
	bedrooms integer,
	receptions integer,
	bathrooms integer,
	rates text,
	heating text,
	epc text,
	status text,
	description text,
	primary key (id)
)
