import { PrismaClient } from "@prisma/client";
import { PrismaPg } from "@prisma/adapter-pg";
import pg from "pg";
import "dotenv/config";

const pool = new pg.Pool({
  connectionString: process.env.DATABASE_URL,
});
const adapter = new PrismaPg(pool, { schema: "chat" });
const prisma = new PrismaClient({ adapter });

// Helper to pick random items from an array
function randomFrom<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)];
}

function randomInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

async function main() {
  console.log("--- Start Seeding (Schema: chat) ---");

  // =============================================
  // 1. CREATE 50 USERS
  // =============================================
  const userSeeds = [
    { username: "SystemBot", email: "system@chatapp.com", status: "online" as const },
    { username: "AdminAlice", email: "alice.admin@chatapp.com", status: "online" as const },
    { username: "DevBob", email: "bob.dev@example.com", status: "away" as const },
    { username: "DesignerCarla", email: "carla@example.com", status: "online" as const },
    { username: "DataDave", email: "dave.data@example.com", status: "dnd" as const },
    { username: "EngEva", email: "eva.eng@example.com", status: "online" as const },
    { username: "FrontendFrank", email: "frank.fe@example.com", status: "offline" as const },
    { username: "GamerGrace", email: "grace.gamer@example.com", status: "online" as const },
    { username: "HackerHank", email: "hank.hack@example.com", status: "away" as const },
    { username: "InternIris", email: "iris.intern@example.com", status: "online" as const },
    { username: "JuniorJake", email: "jake.junior@example.com", status: "offline" as const },
    { username: "KubernetesKara", email: "kara.k8s@example.com", status: "online" as const },
    { username: "LinuxLeo", email: "leo.linux@example.com", status: "dnd" as const },
    { username: "MobileMia", email: "mia.mobile@example.com", status: "online" as const },
    { username: "NetworkNate", email: "nate.net@example.com", status: "away" as const },
    { username: "OpsOlivia", email: "olivia.ops@example.com", status: "online" as const },
    { username: "PythonPete", email: "pete.python@example.com", status: "offline" as const },
    { username: "QAQuinn", email: "quinn.qa@example.com", status: "online" as const },
    { username: "RustRachel", email: "rachel.rust@example.com", status: "dnd" as const },
    { username: "SecuritySam", email: "sam.sec@example.com", status: "online" as const },
    { username: "TechTina", email: "tina.tech@example.com", status: "away" as const },
    { username: "UXUmar", email: "umar.ux@example.com", status: "online" as const },
    { username: "VimVictor", email: "victor.vim@example.com", status: "offline" as const },
    { username: "WebWendy", email: "wendy.web@example.com", status: "online" as const },
    { username: "XMLXander", email: "xander.xml@example.com", status: "away" as const },
    { username: "YAMLYara", email: "yara.yaml@example.com", status: "online" as const },
    { username: "ZshZara", email: "zara.zsh@example.com", status: "dnd" as const },
    { username: "GoGordon", email: "gordon.go@example.com", status: "online" as const },
    { username: "DockerDan", email: "dan.docker@example.com", status: "offline" as const },
    { username: "CloudChloe", email: "chloe.cloud@example.com", status: "online" as const },
    { username: "APIAaron", email: "aaron.api@example.com", status: "away" as const },
    { username: "BashBella", email: "bella.bash@example.com", status: "online" as const },
    { username: "CICDCarlos", email: "carlos.cicd@example.com", status: "dnd" as const },
    { username: "DebugDiana", email: "diana.debug@example.com", status: "online" as const },
    { username: "ElasticEthan", email: "ethan.elastic@example.com", status: "offline" as const },
    { username: "FirewallFiona", email: "fiona.fw@example.com", status: "online" as const },
    { username: "GitGavin", email: "gavin.git@example.com", status: "away" as const },
    { username: "HTTPHolly", email: "holly.http@example.com", status: "online" as const },
    { username: "IoTIvan", email: "ivan.iot@example.com", status: "dnd" as const },
    { username: "JavaJasmine", email: "jasmine.java@example.com", status: "online" as const },
    { username: "KafkaKevin", email: "kevin.kafka@example.com", status: "offline" as const },
    { username: "LambdaLuna", email: "luna.lambda@example.com", status: "online" as const },
    { username: "MongoMarco", email: "marco.mongo@example.com", status: "away" as const },
    { username: "NginxNora", email: "nora.nginx@example.com", status: "online" as const },
    { username: "OAuth_Oscar", email: "oscar.oauth@example.com", status: "dnd" as const },
    { username: "PostgresPaul", email: "paul.pg@example.com", status: "online" as const },
    { username: "RedisRita", email: "rita.redis@example.com", status: "offline" as const },
    { username: "SQLSophia", email: "sophia.sql@example.com", status: "online" as const },
    { username: "TypeScriptTom", email: "tom.ts@example.com", status: "away" as const },
    { username: "UnixUrsula", email: "ursula.unix@example.com", status: "online" as const },
  ];

  const users: any[] = [];
  for (const u of userSeeds) {
    const user = await prisma.users.upsert({
      where: { email: u.email },
      update: {},
      create: u,
    });
    users.push(user);
  }
  console.log(`✅ Created ${users.length} users`);

  // =============================================
  // 2. CREATE 10 ROOMS
  // =============================================
  const roomSeeds = [
    { name: "General Chat", description: "Welcome to the main lobby! Say hi to everyone.", ownerIdx: 0, isPrivate: false },
    { name: "Backend Dev", description: "Go, Rust, Python — all backend talk goes here.", ownerIdx: 27, isPrivate: false },
    { name: "Frontend Lounge", description: "React, Vue, Svelte — pixels and components.", ownerIdx: 6, isPrivate: false },
    { name: "DevOps & Infra", description: "Docker, K8s, CI/CD pipelines and cloud stuff.", ownerIdx: 11, isPrivate: false },
    { name: "Gaming Corner", description: "Off-topic gaming discussions and LFG.", ownerIdx: 7, isPrivate: false },
    { name: "Design & UX", description: "Figma files, design systems, and UX research.", ownerIdx: 3, isPrivate: false },
    { name: "Security Alerts", description: "CVEs, patches, and security best practices.", ownerIdx: 19, isPrivate: true },
    { name: "Random", description: "Memes, random thoughts, and water cooler chat.", ownerIdx: 1, isPrivate: false },
    { name: "Hiring & Careers", description: "Job postings, resume reviews, and career advice.", ownerIdx: 1, isPrivate: false },
    { name: "Book Club", description: "Monthly tech book discussions.", ownerIdx: 20, isPrivate: true },
  ];

  const rooms = [];
  for (const r of roomSeeds) {
    const room = await prisma.rooms.create({
      data: {
        name: r.name,
        description: r.description,
        is_private: r.isPrivate,
        owner_id: users[r.ownerIdx].id,
      },
    });
    rooms.push(room);
  }
  console.log(`✅ Created ${rooms.length} rooms`);

  // =============================================
  // 3. ASSIGN MEMBERS TO ROOMS
  // =============================================
  let memberCount = 0;
  for (let ri = 0; ri < rooms.length; ri++) {
    const room = rooms[ri];
    const ownerIdx = roomSeeds[ri].ownerIdx;

    // Owner is always a member
    const memberIndices = new Set<number>([ownerIdx]);

    // Add 10-30 random members per room
    const targetSize = randomInt(10, 30);
    while (memberIndices.size < targetSize && memberIndices.size < users.length) {
      memberIndices.add(randomInt(0, users.length - 1));
    }

    const memberData = Array.from(memberIndices).map((idx) => ({
      room_id: room.id,
      user_id: users[idx].id,
      role: idx === ownerIdx ? "owner" : randomFrom(["member", "member", "member", "moderator", "admin"]) as string,
    }));

    await prisma.room_members.createMany({ data: memberData });
    memberCount += memberData.length;
  }
  console.log(`✅ Created ${memberCount} room memberships`);

  // =============================================
  // 4. GENERATE MESSAGES (300+)
  // =============================================
  const messageTemplates = [
    "Hey everyone! 👋",
    "Good morning!",
    "Anyone working on something cool today?",
    "Just pushed a huge commit 🚀",
    "Can someone review my PR?",
    "Has anyone tried the new release?",
    "I'm stuck on a bug, any ideas?",
    "Let's schedule a meeting for this.",
    "Great job on the last sprint!",
    "Who broke the build? 😅",
    "lol that's hilarious",
    "I completely agree with that approach.",
    "We should refactor this module.",
    "The docs are really outdated...",
    "I wrote a small utility for this, check it out.",
    "That's a really clean solution!",
    "Is the API down for anyone else?",
    "CI/CD pipeline is green again ✅",
    "Just deployed to staging.",
    "Hotfix incoming...",
    "Can we add more tests for this?",
    "The latency looks much better now.",
    "Who's handling on-call this week?",
    "New RFC is up for review.",
    "I love this framework honestly.",
    "Switching to Vim was the best decision.",
    "VS Code > everything, fight me.",
    "Tabs vs spaces... here we go again.",
    "Dark mode is the only mode.",
    "Did anyone see the tech talk yesterday?",
    "Let's do a knowledge sharing session.",
    "The database migration went smoothly!",
    "We need to add rate limiting.",
    "WebSocket connection keeps dropping.",
    "Just finished the auth module 🔐",
    "Redis cache is saving us so much time.",
    "Should we switch to gRPC?",
    "REST is fine for our use case.",
    "GraphQL would be overkill here.",
    "Microservices or monolith for this?",
    "The monitoring dashboard looks great.",
    "Alert fatigue is real...",
    "Let's set up proper log aggregation.",
    "Anyone going to the conference next month?",
    "I just learned about this Go pattern, it's amazing.",
    "TypeScript generics are so powerful.",
    "Rust borrow checker humbled me today.",
    "Python is still king for scripting.",
    "Docker compose makes life so easy.",
    "K8s is complex but worth it.",
    "Terraform plan looks clean.",
    "Just set up Prometheus + Grafana.",
    "The design mockups look fantastic!",
    "Can we do a UX review session?",
    "Accessibility is not optional.",
    "The color palette needs work.",
    "Mobile responsiveness is broken on Safari.",
    "Let's add dark mode support.",
    "The loading animation is smooth 👌",
    "I'm reading 'Designing Data-Intensive Applications'",
    "Clean Architecture is a must-read.",
    "Has anyone read 'The Phoenix Project'?",
    "Game night this Friday? 🎮",
    "What's everyone playing right now?",
    "Just hit Diamond rank!",
    "The new update ruined the meta.",
    "LFG for ranked matches tonight.",
    "brb, grabbing coffee ☕",
    "Happy Friday everyone! 🎉",
    "Monday meetings should be illegal.",
    "Working from home is the best.",
    "The office snacks are gone again...",
    "Who wants to pair program?",
    "Code review comments are in.",
    "LGTM, ship it! 🚢",
    "Needs more unit tests before merging.",
    "The integration tests are flaky again.",
    "E2E tests are passing now!",
    "Performance benchmarks look promising.",
    "Memory leak found and fixed 🔧",
    "The garbage collector is our friend.",
    "Concurrency bugs are the worst.",
    "Race condition detected in production.",
    "Rollback initiated, investigating now.",
    "Post-mortem scheduled for tomorrow.",
    "Incident resolved, all systems nominal ✅",
    "SLA was maintained, barely 😅",
    "Backup verification completed successfully.",
    "Disaster recovery drill next week.",
    "Security audit passed with flying colors!",
    "Penetration test results are in.",
    "We need to rotate the API keys.",
    "MFA is now mandatory for all accounts.",
    "SSL certificate renewed.",
    "New CVE dropped, patching now.",
    "Dependency update broke everything.",
    "Pinned the version, crisis averted.",
    "npm audit found 47 vulnerabilities 😱",
    "Go modules are so much cleaner.",
    "Just containerized the legacy app.",
    "The migration is 80% complete.",
  ];

  const allMessages = [];

  // Get room members for each room to ensure messages come from actual members
  for (let ri = 0; ri < rooms.length; ri++) {
    const room = rooms[ri];
    const members = await prisma.room_members.findMany({
      where: { room_id: room.id },
      select: { user_id: true },
    });
    const memberUserIds = members.map((m: { user_id: any; }) => m.user_id);

    // Generate 30-60 messages per room
    const msgCount = randomInt(30, 60);
    const baseTime = new Date("2025-12-01T08:00:00Z").getTime();

    for (let i = 0; i < msgCount; i++) {
      // Spread messages over time (every 5-60 minutes)
      const offset = i * randomInt(5, 60) * 60 * 1000;
      const createdAt = new Date(baseTime + offset);

      allMessages.push({
        room_id: room.id,
        user_id: randomFrom(memberUserIds),
        content: randomFrom(messageTemplates),
        message_type: i === 0 ? "system" : randomFrom(["text", "text", "text", "text", "text", "image"]) as string,
        is_edited: Math.random() < 0.05,
        created_at: createdAt,
        updated_at: createdAt,
      });
    }
  }

  // Batch insert messages
  await prisma.messages.createMany({ data: allMessages });
  console.log(`Created ${allMessages.length} messages`);

  // =============================================
  // 5. GENERATE DIRECT MESSAGES (100+)
  // =============================================
  const dmTemplates = [
    "Hey, are you free for a quick call?",
    "Can you take a look at this PR?",
    "Thanks for the help earlier!",
    "Sure, I'll get that done by EOD.",
    "Let me know when you're available.",
    "Did you see the announcement?",
    "The deploy looks good",
    "I'll send you the docs.",
    "Can we sync on this tomorrow?",
    "Happy birthday!",
    "Welcome to the team!",
    "Your presentation was great!",
    "I have a question about the architecture.",
    "Can you share the credentials securely?",
    "Meeting moved to 3pm.",
    "Running late, start without me.",
    "That bug you found was critical, nice catch!",
    "Do you have access to the staging env?",
    "I'll pair with you on that after lunch.",
    "The client loved the demo!",
  ];

  const directMessages = [];
  const dmBaseTime = new Date("2025-12-15T09:00:00Z").getTime();

  for (let i = 0; i < 120; i++) {
    const senderIdx = randomInt(0, users.length - 1);
    let receiverIdx = randomInt(0, users.length - 1);
    while (receiverIdx === senderIdx) {
      receiverIdx = randomInt(0, users.length - 1);
    }

    const offset = i * randomInt(10, 90) * 60 * 1000;
    const createdAt = new Date(dmBaseTime + offset);

    directMessages.push({
      sender_id: users[senderIdx].id,
      receiver_id: users[receiverIdx].id,
      content: randomFrom(dmTemplates),
      message_type: "text" as string,
      is_read: Math.random() < 0.7,
      created_at: createdAt,
    });
  }

  await prisma.direct_messages.createMany({ data: directMessages });
  console.log(`Created ${directMessages.length} direct messages`);

  // =============================================
  // 6. GENERATE ROOM INVITATIONS
  // =============================================
  const invitations = [];
  for (let i = 0; i < 15; i++) {
    const roomIdx = randomInt(0, rooms.length - 1);
    const inviterIdx = roomSeeds[roomIdx].ownerIdx;
    const code = `INV-${Math.random().toString(36).substring(2, 10).toUpperCase()}`;
    const createdAt = new Date(dmBaseTime + i * 3600000);

    invitations.push({
      room_id: rooms[roomIdx].id,
      invited_by: users[inviterIdx].id,
      invite_code: code,
      max_uses: randomFrom([null, 5, 10, 25, 50]),
      use_count: randomInt(0, 5),
      expires_at: Math.random() < 0.5 ? new Date("2026-06-01T00:00:00Z") : null,
      created_at: createdAt,
    });
  }

  await prisma.room_invitations.createMany({ data: invitations });
  console.log(`Created ${invitations.length} room invitations`);

  console.log("\n--- Seeding Finished Successfully --- ");
}

main()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
