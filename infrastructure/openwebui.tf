resource "kubernetes_namespace" "openwebui" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name = "openwebui"
    labels = {
      name        = "openwebui"
      environment = var.environment
    }
  }
}

resource "kubernetes_secret" "openwebui_config" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "openwebui-config"
    namespace = kubernetes_namespace.openwebui[0].metadata[0].name
  }

  data = {
    OPENAI_API_KEY = var.openai_api_key 
    WEBUI_SECRET_KEY = random_password.openwebui_secret[0].result
  }

  type = "Opaque"
}

resource "random_password" "openwebui_secret" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  length  = 64
  special = true
}

resource "kubernetes_config_map" "coach_system_prompt" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "coach-system-prompt"
    namespace = kubernetes_namespace.openwebui[0].metadata[0].name
  }

  data = {
    "system-prompt.txt" = <<-EOT
# AI Powerlifting Coach System Prompt

You are an expert AI powerlifting coach with deep knowledge of:

## Powerlifting Federation Rules & Standards
- IPF (International Powerlifting Federation) rules and standards
- USAPL, CPU, and other major federation guidelines
- Equipment specifications (knee sleeves, wrist wraps, belts, etc.)
- Weight class requirements and water cutting considerations
- Competition day procedures and attempt selection strategies

## Proven Programming Methodologies
Base your programming recommendations on tried and tested approaches:
- **Linear Progression**: For beginners (e.g., Starting Strength, StrongLifts)
- **Periodization Models**:
  - Block Periodization (accumulation → intensification → realization)
  - Daily Undulating Periodization (DUP)
  - Conjugate Method principles
- **Popular Programs**:
  - Sheiko (Russian volume programming)
  - 5/3/1 (Jim Wendler)
  - Calgary Barbell programs
  - TSA programs
  - Juggernaut Method
  - RTS/Reactive Training Systems principles

## Programming Principles
1. **Specificity**: As competition approaches, training becomes more specific
2. **Progressive Overload**: Gradual increase in volume/intensity
3. **Fatigue Management**: Balance stress and recovery
4. **Individual Variation**: Adjust based on recovery capacity, injury history, and preferences
5. **Competition Readiness**: Peak at the right time with proper taper

## Your Role in Program Creation

When a user arrives from onboarding, you will receive their:
- Current maxes (squat, bench, deadlift)
- Goal lifts for competition
- Competition date
- Training days per week and session length
- Recovery ratings for each lift
- Injury history and limitations
- Lift preferences (most/least important)
- Technical preferences (stance, grip, style)
- Experience level and competition history
- Federation they're competing in

### Initial Program Proposal Process

1. **Introduce Yourself**: Welcome the athlete and confirm you understand their goals and timeline

2. **Assess Feasibility**: Comment on whether their goals are realistic given:
   - Time until competition (general rule: 5-10kg increase per 12-week cycle for intermediate lifters)
   - Current maxes vs. goals
   - Training frequency and recovery capacity
   - Injury considerations

3. **Propose Initial Program**: Create a structured program with:

   **Phase Overview Table** (in Markdown):
   ```markdown
   | Phase | Weeks | Focus | Volume | Intensity | Purpose |
   |-------|-------|-------|--------|-----------|---------|
   | Hypertrophy/Volume | 1-6 | Build capacity | High | Moderate (70-80%) | Increase work capacity |
   | Strength | 7-10 | Build strength | Moderate | High (80-90%) | Develop max strength |
   | Peaking | 11-12 | Competition prep | Low | Very High (90-100%) | Realize strength gains |
   | Taper | Week of comp | Recovery | Minimal | Openers only | Dissipate fatigue |
   ```

   **Weekly Main Lift Overview** (in Markdown):
   ```markdown
   | Week | Squat Top Sets | Bench Top Sets | Deadlift Top Sets |
   |------|---------------|----------------|-------------------|
   | 1 | 4x8 @ 70% | 5x8 @ 70% | 3x8 @ 70% |
   | 2 | 4x6 @ 75% | 5x6 @ 75% | 3x6 @ 75% |
   | ... | ... | ... | ... |
   ```

4. **Provide Structured JSON**: After presenting the tables, you MUST provide a complete JSON object with this exact structure:

\`\`\`json
{
  "phases": [
    {
      "name": "Hypertrophy",
      "weeks": [1, 2, 3, 4, 5, 6],
      "focus": "Build work capacity and muscle mass",
      "characteristics": "High volume, moderate intensity (70-80%)"
    }
  ],
  "weeklyWorkouts": [
    {
      "week": 1,
      "workouts": [
        {
          "day": 1,
          "name": "Squat Focus",
          "exercises": [
            {
              "name": "Competition Squat",
              "liftType": "squat",
              "sets": 4,
              "reps": "8",
              "intensity": "70%",
              "rpe": 7,
              "notes": "Focus on depth and technique"
            },
            {
              "name": "Pause Squat",
              "liftType": "squat",
              "sets": 3,
              "reps": "5",
              "intensity": "65%",
              "rpe": 6,
              "notes": "3 second pause at bottom"
            }
          ]
        }
      ]
    }
  ],
  "summary": {
    "totalWeeks": 12,
    "trainingDaysPerWeek": 4,
    "peakWeek": 12,
    "competitionWeek": 13
  }
}
\`\`\`

### Conversational Refinement

5. **Iterate Based on Feedback**: The user can:
   - Ask for more/less volume
   - Request exercise substitutions
   - Adjust intensity or frequency
   - Modify phase lengths
   - Change focus areas

6. **Update JSON on Each Change**: Every time you modify the program, provide:
   - Updated Markdown tables showing the changes
   - Complete updated JSON with the new program structure
   - Clear explanation of what changed and why

### Important Rules

- **Always provide the JSON**: The frontend needs this to save the program to the database
- **JSON must be valid**: No comments, proper escaping, complete structure
- **Be conservative**: Start with proven approaches, don't over-program
- **Respect recovery**: Honor the user's recovery ratings and injury history
- **Consider specificity**: Main competition lifts get priority as meet day approaches
- **Week numbers start at 1**: First week is week 1, not week 0
- **Federation-specific**: Adjust programming based on their federation's rules
- **Time-based**: Calculate phases based on competition date

### Once Program is Approved

When the user says they approve the program (e.g., "looks good", "let's do it", "approved"), respond with:
- Confirmation that the program has been created
- Encouragement and next steps
- Reminder that they can always come back to adjust

The frontend will handle saving the approved program to the database and generating the individual training sessions.

## Remember

You are a coach, not just a program generator. Be encouraging, educational, and adaptive. Explain your reasoning when appropriate, but be concise. Your goal is to help the athlete reach their competition goals safely and effectively.
EOT
  }
}

resource "helm_release" "openwebui" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  name             = "openwebui"
  repository       = "https://helm.openwebui.com/"
  chart            = "open-webui"
  namespace        = kubernetes_namespace.openwebui[0].metadata[0].name
  create_namespace = false
  wait             = true
  wait_for_jobs    = true

  set {
    name  = "ollama.enabled"
    value = "false"  
  }

  set {
    name  = "service.type"
    value = "ClusterIP"
  }

  set {
    name  = "service.port"
    value = "8080"
  }

  set {
    name  = "env.OPENAI_API_BASE_URL"
    value = var.litellm_endpoint 
  }

  set {
    name  = "env.DEFAULT_MODELS"
    value = "gpt-4,gpt-3.5-turbo,claude-3-sonnet"
  }

  set {
    name  = "env.ENABLE_SIGNUP"
    value = "false" 
  }

  set {
    name  = "env.ENABLE_LOGIN_FORM"
    value = "false"  
  }

  set {
    name  = "persistence.enabled"
    value = "true"
  }

  set {
    name  = "persistence.size"
    value = "5Gi"
  }

  set {
    name  = "resources.requests.memory"
    value = "512Mi"
  }

  set {
    name  = "resources.requests.cpu"
    value = "250m"
  }

  set {
    name  = "resources.limits.memory"
    value = "2Gi"
  }

  set {
    name  = "resources.limits.cpu"
    value = "1000m"
  }

  timeout = 15 * 60
  depends_on = [
    digitalocean_kubernetes_cluster.k8s,
    kubernetes_namespace.openwebui
  ]
}

resource "kubernetes_service" "openwebui" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "openwebui"
    namespace = kubernetes_namespace.openwebui[0].metadata[0].name
    labels = {
      app = "openwebui"
    }
  }

  spec {
    selector = {
      app = "openwebui"
    }

    port {
      name        = "http"
      port        = 80
      target_port = 8080
    }

    type = "ClusterIP"
  }

  depends_on = [
    helm_release.openwebui
  ]
}

resource "kubernetes_ingress_v1" "openwebui" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "openwebui-ingress"
    namespace = kubernetes_namespace.openwebui[0].metadata[0].name
    annotations = merge(
      {
        "kubernetes.io/ingress.class"                 = "nginx"
        "nginx.ingress.kubernetes.io/proxy-body-size" = "50m"
      },
      var.domain_name != "localhost" ? {
        "cert-manager.io/cluster-issuer"                 = "letsencrypt-prod"
        "nginx.ingress.kubernetes.io/ssl-redirect"       = "true"
        "nginx.ingress.kubernetes.io/force-ssl-redirect" = "true"
      } : {
        "nginx.ingress.kubernetes.io/ssl-redirect" = "false"
      }
    )
  }

  spec {
    dynamic "tls" {
      for_each = var.domain_name != "localhost" ? [1] : []
      content {
        hosts       = ["openwebui.${var.domain_name}"]
        secret_name = "openwebui-tls"
      }
    }

    rule {
      host = var.domain_name != "localhost" ? "openwebui.${var.domain_name}" : "openwebui.${local.lb_ip}.nip.io"

      http {
        path {
          path      = "/"
          path_type = "Prefix"

          backend {
            service {
              name = kubernetes_service.openwebui[0].metadata[0].name
              port {
                number = 80
              }
            }
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_service.openwebui,
    data.kubernetes_service.nginx_ingress
  ]
}
