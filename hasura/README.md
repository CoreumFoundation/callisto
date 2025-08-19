# Hasura Schema and Metadata

## Make Schema Changes to the Database

First, you need to make the desired changes to your database schema. You can do this in a few ways:

- Using a GUI client like DBeaver, TablePlus, or pgAdmin.
- Using a terminal to write and execute SQL DDL statements (e.g., ALTER TABLE, CREATE TABLE).
- Using the Hasura Console's "Data" tab to add/modify columns, tables, or relationships. This is often the simplest method if you are working directly with the console.

After making the changes, your database's schema will be updated, but Hasura's metadata and the local project files will not yet be in sync.

## Track the New Changes in Hasura

Next, you must tell Hasura about the new database changes.

- If you made the changes using the Hasura Console, any new tables or relationships will likely be automatically tracked. If not, go to the "Data" tab, find the new table or column, and click the "Track" button.

- If you made the changes directly in the database (outside of the console), go to the Hasura Console and look for any untracked tables or columns in the "Data" tab. Hasura will usually show a notification to track them.

### Generate a Migration (Optional but Recommended)

For proper version control and CI/CD, you should generate a migration for the database schema changes.

Navigate to your Hasura project directory in the terminal.

Use the hasura migrate create command to generate a migration file. The --from-server flag tells the CLI to look at the database and compare it with the last known state, creating a migration with the new changes.

```bash
hasura migrate create "add_new_column_to_users" --endpoint "http://localhost:8080" --admin-secret="admin" --from-server
```

This will create a new SQL file in the migrations directory with the CREATE TABLE or ALTER TABLE statements.

## Export Hasura Metadata

> Note: Please install Hasura CLI according to [this guide](https://hasura.io/docs/2.0/hasura-cli/install-hasura-cli/).

This is the key step where you export the YAML files that represent your Hasura configuration. This command captures all the changes you've made in the console, including new permissions, relationships, event triggers, and remote schemas.

- In your terminal, navigate to the Hasura project directory.
- Run the metadata export command.

```bash
hasura metadata export --endpoint "http://localhost:8080" --admin-secret="admin"
```

- This command will update the YAML files in your metadata directory to reflect the current state of the Hasura engine. The table definitions and other configurations will be updated accordingly.

## Export the GraphQL Schema

> Note: Please install graphql(gq) according to [this guide](https://hasura.io/docs/2.0/schema/common-patterns/export-graphql-schema/).

Finally, you can export the latest GraphQL schema. This is not strictly part of the metadata but is useful for client-side tooling and documentation.

```bash
gq http://localhost:8080/v1/graphql -H 'X-Hasura-Admin-Secret: admin' --introspect > schema.graphql
```

This command downloads the complete GraphQL schema from your Hasura instance and saves it to a file named `schema.graphql`. This file is often used for client-side GraphQL code generation.
