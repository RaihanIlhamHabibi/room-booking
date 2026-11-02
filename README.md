# Common Room: full-stack intern take-home

Common Room is a small, fictional campus meeting-room application. Members can sign in as a demo person, browse rooms and reservations, and update a room description. They currently need a coordinator to arrange new meetings. Extend the application so members can find and reserve a room themselves. Your work should fit into the existing frontend, gateway, Go API, and PostgreSQL database.

## Required work

1. **Create a booking.** A signed-in member can choose a room, date, start time, end time, and short purpose. Save valid bookings for the current member. Prevent a room from being booked for conflicting times. Show useful feedback when input or a request fails. After a successful booking, the room's reservations show it, and it remains after an ordinary restart.
2. **Find an available room.** Given a proposed date, start time, end time, and total attendee count, show rooms that can hold the group and are free at that time. The member can proceed from a result to booking that room without reconstructing the whole request.

The total attendee count includes the organizer. Both parts require meaningful frontend and backend work. You may change the interface, API, database, and supporting code as you judge appropriate. The existing UI gives you a visual starting point; there is no prescribed screen design.

## Optional work

**Coordinating attendees:** A member wants to schedule a meeting with a certain duration with several people. Create a scheduling functionality to find possible rooms that can accommodate this meeting: rooms that are large enough to fit all selected people **and** are not occupied when everyone is available. The member should be able to proceed to booking it. The existing application includes fictional people and calendar events for this purpose.

Other useful improvements are welcome. The required work is the priority; extra feature count is not the goal.

## Your approach

Use automated tests for your changes, with meaningful use of mocks. Choose the boundaries and cases that best check your work.

Some details are deliberately open. Make reasonable assumptions and use your judgment to choose an appropriate scope and approach.

You may use documentation, search, and AI tools during the take-home. You are responsible for everything you submit.

Use the [PR description guidance below](#explain-your-work-in-the-pr-description) to explain your approach, verification, and assistance used.

## Submission and interview

You need a GitHub account and a Git client, such as command-line Git or an editor/desktop integration, for the submission workflow.

1. On the starter repository, select **Use this template → Create a new repository**. Create the repository in your own account and choose **Private** visibility. Copy only the default branch; leave **Include all branches** unchecked if shown. **Do not fork the starter.**
2. Invite the designated reviewers as collaborators on your private repository. Reviewer account details will be provided separately. Keep the repository private throughout the assessment.
3. Clone your own private repository. Keep `main` as the supplied starter baseline and create a working branch before making implementation changes.
4. Commit your implementation and tests on that branch, then push it to your own private repository.
5. Open a pull request **within your private repository**, from your working branch into `main`. Leave the PR open and unmerged so reviewers can inspect your changes against the starter.
6. Send the PR URL privately through the submission channel provided to you. Do not open a PR against the public starter or post your submission there.

We will discuss your work in a live interview. If time is limited, prioritize a clear, working required solution. The review will consider the quality of your reasoning, implementation, testing, and communication across both application layers.

### Explain your work in the PR description

Write a concise engineering account that helps a reviewer understand your contribution, your choices, and the evidence behind it. Paragraphs or bullets are both welcome.

- **What changed:** Summarize the required work you completed, any optional work, and anything unfinished.
- **How it fits together:** Include comprehensive diagram(s) showing the implemented application's relevant components and connections—including the frontend, Next.js proxy, Kong gateway, Go API, and PostgreSQL—how booking creation and availability search travel through them, and where components run—including the browser and Docker containers. Choose a format that makes the system clear; individual functions and files do not need to be diagrammed.
- **Why you chose this approach:** Explain the decisions and assumptions that materially affect behavior, correctness, or maintainability, including relevant tradeoffs.
- **Why you believe it works:** Explain how you checked the required behavior. Connect your checks to important implementation risks, describe what you observed, and identify known limitations or remaining uncertainty. Where you used mocks, explain what they establish and what they leave unchecked. If blocked, describe what you tried and what you would do next.
- **Assistance used:** Briefly disclose external or AI assistance and where it helped.
- **Technical Reflection:** Include written answers to all ten [questions below](#technical-reflection) in a clearly labeled Technical Reflection section of the same PR description, alongside the implementation write-up above.

Explain each point once. You may link to supporting code or documentation rather than repeat it.

## Technical Reflection

These questions guide your exploration of the project and the engineering concepts we will discuss in the interview.

**Write your answers to all ten questions in the PR description as part of your submission write-up**, under a Technical Reflection heading. Number your answers to match the questions. These written responses are required alongside the implementation summary, diagrams, decisions, verification, and assistance disclosure described above.

1. **TypeScript:** What is TypeScript, and how does it relate to JavaScript? What benefits and limitations does it bring to developing an application like this one?

2. **Go:** What is an interface in Go, and how does a type satisfy one? Use an example from this project to explain how different implementations can be substituted, including in a test.

3. **React:** Using a component from this project, explain how props, state, and hooks determine what appears on screen and what causes it to update.

4. **Frontend structure and data access:** Trace an existing frontend read operation from a Next.js route through the relevant shared packages to the API client. Explain what Refine contributes along that path.

5. **Backend structure:** Trace an existing backend operation through its Echo handler, service, and repository. Explain their responsibilities and how FX connects their dependencies.

6. **Testing:** Using a test from the starter or your implementation, explain why a dependency is mocked, what the test establishes, and what would require a check using the real dependency.

7. **Kong:** What is an API gateway, and what role does Kong serve here? Explain how a browser API request reaches the Go backend through Next.js and Kong.

8. **Docker:** How do images, containers, and Docker Compose work together to run this application? Explain why saved data survives an ordinary restart and what action removes it.

9. **Git:** What is the relationship between commits, branches, and pull requests? Use your submission to explain the PR’s source and target branches, what changes it presents for review, and what merging it would do.

10. **CI:** What is continuous integration, and why is it useful? Use this repository’s workflow to explain what its automated checks establish and what a successful run does not guarantee.

## Run the starter

You need Docker with Compose v2.23 or later, an editor, and an internet connection for the initial image and dependency downloads. Node.js, Go, pnpm, and Corepack run inside the containers; you do not need to install them on your computer. The Makefile provides optional shortcuts, but the commands below do not require Make.

From this directory:
```sh
docker compose up --build -d
```

Open <http://localhost:13000>. The demo login asks you to select a fictional person; it does not use passwords. The gateway is available at <http://localhost:18080/api/health> and PostgreSQL binds to `127.0.0.1:15432`. These are the default host ports; you can change `WEB_PORT`, `KONG_PORT`, and `POSTGRES_PORT` in a local `.env` file if they are occupied. Use the chosen web/gateway ports in browser URLs. Internal service connections do not change.

The database is initialized once through the `setup` service before the API starts. It creates the schema and example data. Ordinary container restarts preserve later changes, including room edits and candidate-created records. If you add a migration, run `docker compose run --build --rm setup` to apply it. Do not edit already-applied migrations; add a new one.

Useful commands:
```sh
docker compose ps
docker compose logs -f api web
docker compose up --build --watch   # sync source edits; restart Go after edits
docker compose run --build --rm setup       # apply new migrations; retain data
docker compose down                 # stop; retain data
```

To discard all local starter data and restore the fixtures, run the following from this directory. This deletes only this starter project's database volume:
```sh
docker compose down --volumes
docker compose up --build -d
```

The Compose file reads optional values from `.env`; `.env.example` contains fictional local defaults only. You can run the default setup without creating `.env`.

Compose uses this directory's name to group its containers, network, and database volume. If you run multiple copies on the same Docker daemon, use distinct directory names or set a unique `COMPOSE_PROJECT_NAME` in each copy's `.env`, and choose unused host ports. Two directories with the same name do not get separate project identities automatically. On a shared Docker daemon, stop or reset only a project you own. Changing a project name selects a different database volume; it does not move or delete an existing one.

## Explore the project

This is a small application already in progress. Its basic features demonstrate the project's layering:

| Location | Responsibility |
| --- | --- |
| `platforms/web` | Next.js routes and app setup |
| `packages/features` | Domain UI |
| `packages/api` | Browser API client and data access |
| `packages/ui` | Reusable interface elements |
| `backend/cmd`, `backend/internal` | Go entry points, HTTP handlers, services, and repositories |
| `backend/database` | Migrations and fictional fixtures |
| `infra/kong` | Gateway routing |

Browser requests use `/api`, which the Next.js development server forwards to Kong. Kong routes to the Go API; the API reads and writes PostgreSQL. The backend uses Echo, FX, and GORM. The frontend uses TypeScript, React, Next.js, Refine, and Tailwind. The supplied tests show some of the project's testing tools; extend them where they provide useful confidence.

The fixture dates are **May 13–15, 2030**. Times in the application use **Asia/Jakarta**. The optional scheduling task searches within **09:00–17:00** on the chosen date. The fixture dates are intentionally in the future so the starter remains explorable after publication.

## Checks

Run these from this directory. The commands use temporary containers, so local Node.js, Go, pnpm, and Corepack installations are not needed. The application does not need to be running for these code checks:

```sh
# Frontend: TypeScript, automated tests, and production build
docker compose run --build --rm --no-deps web corepack pnpm typecheck
docker compose run --build --rm --no-deps web corepack pnpm test
docker compose run --build --rm --no-deps web corepack pnpm build

# Backend automated tests
docker compose run --build --rm --no-deps api go test ./...
```

`--build` includes your latest source and dependencies. `--rm` removes the temporary container afterward; `--no-deps` avoids starting the database or other services for these checks. Images and downloaded dependencies are cached between builds. These checks do not reset your application data.

The frontend container uses Corepack to select the pnpm version pinned in `package.json`, and the lockfile fixes the dependency versions. This happens inside Docker; no separate package-manager setup is required.

With the application running, open <http://localhost:18080/api/health> to check that Kong can reach the API. Also verify your feature in the browser: passing code checks and a health response do not establish that the full user flow works.

Use the [PR description guidance above](#explain-your-work-in-the-pr-description) to report your verification and any unresolved failures.
