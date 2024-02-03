-- CURRENCIES
insert into
    liquiswiss.go_currencies (
        id, code, description, locale_code
    )
values (
        1, 'CHF', 'Schweizer Franken', 'CH'
    );

insert into liquiswiss.go_currencies (id, code, description, locale_code)
values (2, 'EUR', 'Euro', 'DE');


insert into
    liquiswiss.go_currencies (
        id, code, description, locale_code
    )
values (3, 'USD', 'US Dollar', 'US');

-- CATEGORIES
insert into
    liquiswiss.go_categories (id, name)
values (1, 'Hosting');

insert into
    liquiswiss.go_categories (id, name)
values (2, 'Steuern');

insert into
    liquiswiss.go_categories (id, name)
values (3, 'Lizenzen');

insert into
    liquiswiss.go_categories (id, name)
values (4, 'Sozialabgaben');