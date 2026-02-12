package router

import (
	"backend/controller"
	"backend/middleware"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.Post("/register", controller.Register)
	auth.Post("/login", controller.Login)
	auth.Post("/logout", controller.Logout)

	api.Use(middleware.AllRolesProtected())

	user := api.Group("/user", middleware.UserProtected())

	user.Get("/profile", controller.GetProfile)
	user.Post("channels", controller.CreateChannel)
	user.Post("/channels/:id/messages", controller.SendMessage)

	agent := api.Group("/agent", middleware.AgentProtected())

	agent.Get("/conversations", controller.GetAgentConversations)
	agent.Get("/channels/available", controller.GetAvailableChannels)
	agent.Get("/channels/stats", controller.GetChannelStats)
	agent.Get("/channels/:id", controller.GetChannelByID)

	agent.Patch("/channels/:id/assign", controller.AssignChannel)
	agent.Post("/channels/:id/close", controller.CloseChannel)
	agent.Post("/channels/:id/messages", controller.SendMessage)

	admin := api.Group("/admin", middleware.AdminProtected())

	admin.Get("/users", controller.GetAllUsers)
	admin.Get("/channels/available", controller.GetAvailableChannels)
	admin.Patch("/channels/:id/assign", controller.AssignChannel)

	adminOrAgent := api.Group("/conversations", middleware.AdminOrAgentProtected())
	adminOrAgent.Get("/:id", controller.GetChannelByID)

}
