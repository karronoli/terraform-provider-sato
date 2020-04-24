package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Coffee -
type Coffee struct {
	ID          int                 `db:"id" json:"id"`
	Name        string              `db:"name" json:"name"`
	Teaser      string              `db:"teaser" json:"teaser"`
	Description string              `db:"description" json:"description"`
	Price       float64             `db:"price" json:"price"`
	Image       string              `db:"image" json:"image"`
	Ingredients []CoffeeIngredients `json:"ingredients"`
}

// CoffeeIngredients -
type CoffeeIngredients struct {
	ID int `json:"id"`
}

func dataSourceCoffees() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCoffeesRead,
		Schema: map[string]*schema.Schema{
			"coffees": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"teaser": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"price": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"image": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"ingredients": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ingredient_id": &schema.Schema{
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceCoffeesRead(d *schema.ResourceData, m interface{}) error {
	var client = &http.Client{Timeout: 10 * time.Second}

	r, err := client.Get("http://localhost:9090/coffees")
	if err != nil {
		return err
	}
	defer r.Body.Close()
	coffees := make([]map[string]interface{}, 0)
	err = json.NewDecoder(r.Body).Decode(&coffees)
	if err != nil {
		return err
	}

	if err := d.Set("coffees", coffees); err != nil {
		return err
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}