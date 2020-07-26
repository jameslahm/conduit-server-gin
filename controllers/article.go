package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/jameslahm/conduit-server-gin/middlewares"
	"github.com/jameslahm/conduit-server-gin/models"
	"github.com/jameslahm/conduit-server-gin/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetArticlesArgs args for get articles
type GetArticlesArgs struct {
	Tag       string `form:"tag"`
	Author    string `form:"author"`
	Favorited string `form:"favorited"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
}

// GetAllArticles get all articles
// @summary Get All Articles
// @description get all articles using filter
// @tags Article
// @accept json
// @produce json
// @param limit query string false "limit nums of articles"
// @param offset query string false "offset of articles"
// @param tag query string false "tag of articles"
// @param author query string false "author of articles"
// @param favorited query string false "articles favorted by"
// @router /articles [get]
// @success 200 {array} models.Article
func GetAllArticles(c *gin.Context) {
	var args GetArticlesArgs
	args.Limit = 20
	args.Offset = 0

	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	articleCollection := client.Database("conduit").Collection("articles")
	userCollection := client.Database("conduit").Collection("users")

	var query bson.D = bson.D{}
	var author models.User
	if args.Author != "" {
		err := userCollection.FindOne(ctx, bson.M{"username": args.Author}).Decode(&author)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		query = append(query, primitive.E{Key: "author", Value: author.ID})
	}
	if args.Tag != "" {
		query = append(query, primitive.E{Key: "tagList", Value: bson.D{{Key: "$in", Value: []string{args.Tag}}}})
	}
	if args.Favorited != "" {
		err := userCollection.FindOne(ctx, bson.M{"username": args.Author}).Decode(&author)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		query = append(query, primitive.E{Key: "_id", Value: bson.D{{Key: "$in", Value: author.Favorites}}})
	}

	matchStage := bson.D{{Key: "$match", Value: query}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "users"}, {Key: "localField", Value: "author"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "author"}}}}
	unwindStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$author"}, {Key: "preserveNullAndEmptyArrays", Value: false}}}}
	skipStage := bson.D{{Key: "$skip", Value: args.Offset}}
	limitStage := bson.D{{Key: "$limit", Value: args.Limit}}
	cursor, err := articleCollection.Aggregate(ctx, mongo.Pipeline{matchStage, lookupStage, unwindStage, skipStage, limitStage})
	var articles []models.ArticleWithAuthor
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	err = cursor.All(ctx, &articles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	counts, err := articleCollection.CountDocuments(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var loginUser models.User
	claims, err := middlewares.Authenticate(c)
	if err == nil {
		id, err := primitive.ObjectIDFromHex(claims.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		err = userCollection.FindOne(ctx, bson.M{
			"_id": id,
		}).Decode(&loginUser)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	var articlesJSON []models.ArticleJSON = make([]models.ArticleJSON, len(articles))
	for i, article := range articles {
		articlesJSON[i].ArticleBase = article.ArticleBase
		articlesJSON[i].Author = article.Author.ToProfile(&loginUser)
	}

	c.JSON(http.StatusOK, gin.H{
		"articles":      articles,
		"articlesCount": counts,
	})
}

// GetFeedArgs get feed articles
type GetFeedArgs struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

// GetFeedArticles get feed articles
func GetFeedArticles(c *gin.Context) {
	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	articleCollection := client.Database("conduit").Collection("articles")
	userCollection := client.Database("conduit").Collection("users")

	var loginUser models.User
	claims, err := middlewares.Authenticate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	id, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	err = userCollection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&loginUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	var args GetFeedArgs
	args.Limit = 20
	args.Offset = 0

	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var query bson.M = make(primitive.M)
	query["author"] = bson.M{
		"$in": loginUser.Following,
	}

	matchStage := bson.D{{Key: "$match", Value: query}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "users"}, {Key: "localField", Value: "author"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "author"}}}}
	unwindStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$author"}, {Key: "preserveNullAndEmptyArrays", Value: false}}}}
	skipStage := bson.D{{Key: "$skip", Value: args.Offset}}
	limitStage := bson.D{{Key: "$limit", Value: args.Limit}}
	cursor, err := articleCollection.Aggregate(ctx, mongo.Pipeline{matchStage, lookupStage, unwindStage, skipStage, limitStage})
	counts, err := articleCollection.CountDocuments(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	var articles []models.ArticleWithAuthor
	err = cursor.All(ctx, &articles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var articlesJSON []models.ArticleJSON = make([]models.ArticleJSON, len(articles))
	for i, article := range articles {
		articlesJSON[i].ArticleBase = article.ArticleBase
		articlesJSON[i].Author = article.Author.ToProfile(&loginUser)
	}

	c.JSON(http.StatusOK, gin.H{
		"articles":      articlesJSON,
		"articlesCount": counts,
	})
}

// GetArticle get single article
func GetArticle(c *gin.Context) {
	var loginUser models.User
	claims, err := middlewares.Authenticate(c)

	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")
	articleCollection := client.Database("conduit").Collection("articles")

	id, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	err = userCollection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&loginUser)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	slug := c.Param("slug")

	var article models.ArticleWithAuthor
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "slug", Value: slug}}}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "users"}, {Key: "localField", Value: "author"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "author"}}}}
	cursor, err := articleCollection.Aggregate(ctx, mongo.Pipeline{matchStage, lookupStage})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	for cursor.Next(ctx) {
		err := cursor.Decode(&article)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		break
	}
	var articleJSON = models.ArticleJSON{ArticleBase: article.ArticleBase, Author: article.Author.ToProfile(&loginUser)}
	c.JSON(http.StatusOK, gin.H{
		"article": articleJSON,
	})
}

// CreateArticleInput create article post data
type CreateArticleInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagList     []string `json:"tagList"`
}

// CreateArticle create article
func CreateArticle(c *gin.Context) {
	claims, err := middlewares.Authenticate(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	id, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")
	articleCollection := client.Database("conduit").Collection("articles")

	var loginUser models.User
	err = userCollection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&loginUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	var data CreateArticleInput
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var article models.Article
	article.Title = data.Title
	article.Description = data.Description
	article.Body = data.Body
	article.TagList = data.TagList
	article.Author = loginUser.ID
	article.Slug = slug.Make(data.Title)

	_, err = articleCollection.InsertOne(ctx, article)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}

	var articleJSON models.ArticleJSON
	articleJSON.ArticleBase = article.ArticleBase
	articleJSON.Author = loginUser.ToProfile(nil)
	c.JSON(http.StatusOK, gin.H{
		"article": articleJSON,
	})
}

// UpdateArticleInput update article data
type UpdateArticleInput = CreateArticleInput

// UpdateArticle update article
func UpdateArticle(c *gin.Context) {
	claims, err := middlewares.Authenticate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	id, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")
	articleCollection := client.Database("conduit").Collection("articles")

	var loginUser models.User
	err = userCollection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&loginUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	var data UpdateArticleInput
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var article models.Article
	article.Title = data.Title
	article.Description = data.Description
	article.Body = data.Body
	article.TagList = data.TagList

	_, err = articleCollection.UpdateOne(ctx, bson.M{
		"slug":   c.Param("slug"),
		"author": loginUser.ID,
	}, bson.M{
		"$set": article,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var articleJSON models.ArticleJSON
	err = articleCollection.FindOne(ctx, bson.M{"slug": c.Param("slug")}).Decode(&article)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	articleJSON.ArticleBase = article.ArticleBase
	articleJSON.Author = loginUser.ToProfile(nil)

	c.JSON(http.StatusOK, gin.H{
		"article": articleJSON,
	})

}

// DeleteArticle delete article
func DeleteArticle(c *gin.Context) {
	claims, err := middlewares.Authenticate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	id, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")
	articleCollection := client.Database("conduit").Collection("articles")

	var loginUser models.User
	err = userCollection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&loginUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	var data UpdateArticleInput
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var article models.Article
	article.Title = data.Title
	article.Description = data.Description
	article.Body = data.Body
	article.TagList = data.TagList

	_, err = articleCollection.DeleteOne(ctx, bson.M{
		"slug":   c.Param("slug"),
		"author": loginUser.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// FavoriteArticle favorite article
func FavoriteArticle(c *gin.Context) {
	claims, err := middlewares.Authenticate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	id, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")
	articleCollection := client.Database("conduit").Collection("articles")

	var loginUser models.User
	err = userCollection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&loginUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	var article models.Article
	err = articleCollection.FindOne(ctx, bson.M{
		"slug": c.Param("slug"),
	}).Decode(&article)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = loginUser.Favorite(&article)
	if err == nil {
		_, err = userCollection.UpdateOne(ctx, bson.M{
			"_id": loginUser.ID,
		}, loginUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		article.FavoritesCount++
		_, err = articleCollection.UpdateOne(ctx, bson.M{
			"_id": article.ID,
		}, article)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	var articleJSON models.ArticleJSON
	articleJSON.ArticleBase = article.ArticleBase
	articleJSON.Author = loginUser.ToProfile(nil)
	c.JSON(http.StatusOK, gin.H{
		"article": articleJSON,
	})

}

// UnFavoriteArticle unfavorite article
func UnFavoriteArticle(c *gin.Context) {
	claims, err := middlewares.Authenticate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	id, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	client, ctx, cancel := utils.GetConnection()
	defer cancel()
	userCollection := client.Database("conduit").Collection("users")
	articleCollection := client.Database("conduit").Collection("articles")

	var loginUser models.User
	err = userCollection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&loginUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	var article models.Article
	err = articleCollection.FindOne(ctx, bson.M{
		"slug": c.Param("slug"),
	}).Decode(&article)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = loginUser.UnFavorite(&article)
	if err == nil {
		_, err = userCollection.UpdateOne(ctx, bson.M{
			"_id": loginUser.ID,
		}, loginUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		article.FavoritesCount--
		_, err = articleCollection.UpdateOne(ctx, bson.M{
			"_id": article.ID,
		}, article)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	var articleJSON models.ArticleJSON
	articleJSON.ArticleBase = article.ArticleBase
	articleJSON.Author = loginUser.ToProfile(nil)
	c.JSON(http.StatusOK, gin.H{
		"article": articleJSON,
	})

}

// GetTags get tas
func GetTags(c *gin.Context) {
	client, ctx, cancel := utils.GetConnection()
	defer cancel()

	articleCollection := client.Database("conduit").Collection("articles")
	distinctResult, err := articleCollection.Distinct(ctx, "tagList", bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"tags": distinctResult,
	})
}
